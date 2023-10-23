package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudflare/cfssl/log"
	peerproto "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/status"

	"github.com/s7techlab/hlf-sdk-go/api"
	clienterrors "github.com/s7techlab/hlf-sdk-go/client/errors"
)

var ErrEndorsingMSPsRequired = errors.New(`endorsing MSPs required`)

type PeerPool struct {
	ctx    context.Context
	cancel context.CancelFunc
	logger *zap.Logger

	mspPeers map[string][]*peerPoolPeer
	storeMx  sync.RWMutex
}

type peerPoolPeer struct {
	peer  api.Peer
	ready bool
}

type endorseChannelResponse struct {
	Response *peerproto.ProposalResponse
	Error    error
}

func NewPeerPool(ctx context.Context, log *zap.Logger) *PeerPool {
	ctx, cancel := context.WithCancel(ctx)

	return &PeerPool{
		mspPeers: make(map[string][]*peerPoolPeer),
		logger:   log.Named(`peer-pool`),
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (p *PeerPool) GetPeers() map[string][]api.Peer {
	m := make(map[string][]api.Peer, 0)

	for mspId, peers := range p.mspPeers {
		for _, poolPeer := range peers {
			m[mspId] = append(m[mspId], poolPeer.peer)
		}
	}

	return m
}

func (p *PeerPool) GetMSPPeers(mspID string) []api.Peer {
	var peers []api.Peer
	if mspPeers, ok := p.mspPeers[mspID]; ok {
		for _, mspPeer := range mspPeers {
			peers = append(peers, mspPeer.peer)
		}
	}
	return peers
}

func (p *PeerPool) Add(mspId string, peer api.Peer, peerChecker api.PeerPoolCheckStrategy) error {
	p.logger.Debug(`add peer`,
		zap.String(`msp_id`, mspId),
		zap.String(`peerUri`, peer.Uri()))

	p.storeMx.Lock()
	defer p.storeMx.Unlock()

	if peers, ok := p.mspPeers[mspId]; !ok {
		p.mspPeers[mspId] = p.addPeer(peer, make([]*peerPoolPeer, 0), peerChecker)
	} else {
		if !p.isPeerInPool(peer, peers) {
			p.mspPeers[mspId] = p.addPeer(peer, peers, peerChecker)
		}
	}
	return nil
}

func (p *PeerPool) addPeer(peer api.Peer, peerSet []*peerPoolPeer, peerChecker api.PeerPoolCheckStrategy) []*peerPoolPeer {
	pp := &peerPoolPeer{peer: peer, ready: true}
	aliveChan := make(chan bool)
	go peerChecker(p.ctx, peer, aliveChan)
	go p.poolChecker(p.ctx, aliveChan, pp)
	return append(peerSet, pp)
}

func (p *PeerPool) isPeerInPool(peer api.Peer, peerSet []*peerPoolPeer) bool {
	for _, pp := range peerSet {
		if peer.Uri() == pp.peer.Uri() {
			return true
		}
	}

	return false
}

func (p *PeerPool) poolChecker(ctx context.Context, aliveChan chan bool, peer *peerPoolPeer) {
	//log := p.log.Named(`poolChecker`)

	for {
		select {
		case <-ctx.Done():
			//log.Debug(`Context canceled`)
			return

		case alive, ok := <-aliveChan:
			//log.Debug(`Got alive data about peer`, zap.String(`peerUri`, peer.peer.Uri()), zap.Bool(`alive`, alive))
			if !ok {
				return
			}

			if !alive {
				p.logger.Warn(`peer connection is dead`, zap.String(`peerUri`, peer.peer.Uri()))
			}

			p.storeMx.Lock()
			peer.ready = alive
			p.storeMx.Unlock()
		}
	}
}

// EndorseOnMSP finds first ready peer in pool for specified mspId , endorses proposal and returns proposal response
// - no load balancing between msp peers
// - no data is not sent to the orderer
func (p *PeerPool) EndorseOnMSP(ctx context.Context, mspID string, proposal *peerproto.SignedProposal) (*peerproto.ProposalResponse, error) {
	p.storeMx.RLock()
	//check MspId exists
	peers, exists := p.mspPeers[mspID]
	p.storeMx.RUnlock()

	if !exists {
		return nil, fmt.Errorf(`msp_id=%s: %w`, mspID, ErrMSPNotFound)
	}

	//check peers for MspId exists
	if len(peers) == 0 {
		return nil, fmt.Errorf(`msp_id=%s: %w`, mspID, ErrNoPeersForMSP)
	}

	var lastError error

	for pos, poolPeer := range peers {
		if !poolPeer.ready {
			p.logger.Debug(ErrPeerNotReady.Error(), zap.String(`uri`, poolPeer.peer.Uri()))
			continue
		}

		log.Debug(`Sending endorse to peer...`,
			zap.String(`mspId`, mspID),
			zap.String(`uri`, poolPeer.peer.Uri()),
			zap.Int(`peerPos`, pos),
			zap.Int(`peers in msp pool`, len(peers)))

		propResp, err := poolPeer.peer.Endorse(ctx, proposal)
		if err != nil {
			// GRPC error
			if s, ok := status.FromError(err); ok {
				if s.Code() == codes.Unavailable {
					log.Debug(`peer GRPC unavailable`, zap.String(`mspId`, mspID), zap.String(`peer_uri`, poolPeer.peer.Uri()))
					//poolPeer.ready = false
				} else {
					log.Debug(`unexpected GRPC error code from peer`,
						zap.String(`peer_uri`, poolPeer.peer.Uri()), zap.Uint32(`code`, uint32(s.Code())),
						zap.String(`code_str`, s.Code().String()), zap.Error(s.Err()))
					// not mark as not ready
				}

				// next mspId peer
				lastError = fmt.Errorf("peer %s: %w", poolPeer.peer.Uri(), err)
				continue
			}

			log.Debug(`peer endorsement failed`,
				zap.String(`mspId`, mspID),
				zap.String(`peer_uri`, poolPeer.peer.Uri()),
				zap.String(`error`, err.Error()))

			return propResp, errors.Wrap(err, poolPeer.peer.Uri())
		}

		log.Debug(`endorse complete on peer`, zap.String(`mspId`, mspID), zap.String(`uri`, poolPeer.peer.Uri()))
		return propResp, nil
	}

	if lastError == nil {
		// all peers were not ready
		return nil, clienterrors.ErrNoReadyPeers{MspId: mspID}
	}

	return nil, lastError
}

func (p *PeerPool) EndorseOnMSPs(ctx context.Context, mspIDs []string, proposal *peerproto.SignedProposal) ([]*peerproto.ProposalResponse, error) {
	if len(mspIDs) == 0 {
		return nil, ErrEndorsingMSPsRequired
	}

	respList := make([]*peerproto.ProposalResponse, 0)
	respChan := make(chan endorseChannelResponse)

	// send all proposals concurrently
	for i := 0; i < len(mspIDs); i++ {
		go func(mspId string) {
			resp, err := p.EndorseOnMSP(ctx, mspId, proposal)
			respChan <- endorseChannelResponse{Response: resp, Error: err}
		}(mspIDs[i])
	}

	var errOccurred bool

	mErr := new(clienterrors.MultiError)

	// collecting peer responses
	for i := 0; i < len(mspIDs); i++ {
		resp := <-respChan
		if resp.Error != nil {
			errOccurred = true
			mErr.Add(resp.Error)
		}
		respList = append(respList, resp.Response)
	}

	if errOccurred {
		return respList, mErr
	}

	return respList, nil
}

func (p *PeerPool) DeliverClient(mspId string, identity msp.SigningIdentity) (api.DeliverClient, error) {
	poolPeer, err := p.FirstReadyPeer(mspId)
	if err != nil {
		return nil, err
	}
	return poolPeer.DeliverClient(identity)
}

func (p *PeerPool) FirstReadyPeer(mspId string) (api.Peer, error) {

	p.storeMx.RLock()
	peers, ok := p.mspPeers[mspId]
	p.storeMx.RUnlock()

	if !ok {
		return nil, ErrMSPNotFound
	}

	//check peers for MspId exists
	if len(peers) == 0 {
		log.Error(ErrNoPeersForMSP.Error(), zap.String(`mspId`, mspId))
	}

	for _, poolPeer := range peers {
		if poolPeer.ready {
			return poolPeer.peer, nil
		}
	}

	return nil, clienterrors.ErrNoReadyPeers{MspId: mspId}
}

func (p *PeerPool) Close() error {
	return nil
}

func StrategyGRPC(d time.Duration) api.PeerPoolCheckStrategy {
	return func(ctx context.Context, peer api.Peer, alive chan bool) {
		t := time.NewTicker(d)
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if peer.Conn().GetState() == connectivity.Ready {
					alive <- true
				} else {
					alive <- false
				}
			}
		}
	}
}
