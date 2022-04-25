package pool

import (
	"context"
	"fmt"
	"sync"

	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/s7techlab/hlf-sdk-go/api"
)

type peerPool struct {
	log    *zap.Logger
	ctx    context.Context
	cancel context.CancelFunc

	store   map[string][]*peerPoolPeer
	storeMx sync.RWMutex
}

type peerPoolPeer struct {
	peer  api.Peer
	ready bool
}

func (p *peerPool) GetPeers() map[string][]api.Peer {
	m := make(map[string][]api.Peer, 0)

	for mspId, peerArr := range p.store {
		for _, peerFromArr := range peerArr {
			m[mspId] = append(m[mspId], peerFromArr.peer)
		}
	}

	return m
}

func (p *peerPool) GetMSPPeers(mspID string) []api.Peer {
	var peers []api.Peer
	if mspPeers, ok := p.store[mspID]; ok {
		for _, mspPeer := range mspPeers {
			peers = append(peers, mspPeer.peer)
		}
	}
	return peers
}

func (p *peerPool) Add(mspId string, peer api.Peer, peerChecker api.PeerPoolCheckStrategy) error {
	log := p.log.Named(`Add`).With(zap.String(`mspId`, mspId))
	log.Debug(`add peer`, zap.String(`peerUri`, peer.Uri()))
	p.storeMx.Lock()
	defer p.storeMx.Unlock()

	if peers, ok := p.store[mspId]; !ok {
		p.store[mspId] = p.addPeer(peer, make([]*peerPoolPeer, 0), peerChecker)
	} else {
		if !p.isPeerInPool(peer, peers) {
			p.store[mspId] = p.addPeer(peer, peers, peerChecker)
		}
	}
	return nil
}

func (p *peerPool) addPeer(peer api.Peer, peerSet []*peerPoolPeer, peerChecker api.PeerPoolCheckStrategy) []*peerPoolPeer {
	pp := &peerPoolPeer{peer: peer, ready: true}
	aliveChan := make(chan bool)
	go peerChecker(p.ctx, peer, aliveChan)
	go p.poolChecker(p.ctx, aliveChan, pp)
	return append(peerSet, pp)
}

func (p *peerPool) isPeerInPool(peer api.Peer, peerSet []*peerPoolPeer) bool {
	for _, pp := range peerSet {
		if peer.Uri() == pp.peer.Uri() {
			return true
		}
	}

	return false
}

func (p *peerPool) poolChecker(ctx context.Context, aliveChan chan bool, peer *peerPoolPeer) {
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
				p.log.Warn(`Peer connection is dead`, zap.String(`peerUri`, peer.peer.Uri()))
			}

			p.storeMx.Lock()
			peer.ready = alive
			p.storeMx.Unlock()
		}
	}
}

// Process finds first ready peer in pool for specified mspId , endorses proposal and returns proposal response
// - no load balancing between msp peers
// - no data is not sent to the orderer
func (p *peerPool) Process(ctx context.Context, mspId string, proposal *peer.SignedProposal) (*peer.ProposalResponse, error) {
	log := p.log.Named(`Process`)
	p.storeMx.RLock()

	//check MspId exists
	peers, exists := p.store[mspId]
	p.storeMx.RUnlock()

	if !exists {
		log.Error(api.ErrMSPNotFound.Error(), zap.String(`mspId`, mspId))
		return nil, api.ErrMSPNotFound
	}

	//check peers for MspId exists
	if len(peers) == 0 {
		log.Error(api.ErrNoPeersForMSP.Error(), zap.String(`mspId`, mspId))
	}

	var lastError error

	for pos, poolPeer := range peers {
		if !poolPeer.ready {
			log.Debug(api.ErrPeerNotReady.Error(), zap.String(`uri`, poolPeer.peer.Uri()))
			continue
		}

		log.Debug(`Sending endorse to peer...`,
			zap.String(`mspId`, mspId),
			zap.String(`uri`, poolPeer.peer.Uri()),
			zap.Int(`peerPos`, pos),
			zap.Int(`peers in msp pool`, len(peers)))

		propResp, err := poolPeer.peer.Endorse(ctx, proposal)
		if err != nil {
			// GRPC error
			if s, ok := status.FromError(err); ok {
				if s.Code() == codes.Unavailable {
					log.Debug(`Peer GRPC unavailable`, zap.String(`mspId`, mspId), zap.String(`peer_uri`, poolPeer.peer.Uri()))
					//poolPeer.ready = false
				} else {
					log.Debug(`Unexpected GRPC error code from peer`,
						zap.String(`peer_uri`, poolPeer.peer.Uri()), zap.Uint32(`code`, uint32(s.Code())),
						zap.String(`code_str`, s.Code().String()), zap.Error(s.Err()))
					// not mark as not ready
				}

				// next mspId peer
				lastError = fmt.Errorf("peer %s: %w", poolPeer.peer.Uri(), err)
				continue
			}

			log.Debug(`Peer endorsement failed`, zap.String(`mspId`, mspId), zap.String(`peer_uri`, poolPeer.peer.Uri()), zap.String(`error`, err.Error()))

			return propResp, errors.Wrap(err, poolPeer.peer.Uri())
		}

		log.Debug(`Endorse complete on peer`, zap.String(`mspId`, mspId), zap.String(`uri`, poolPeer.peer.Uri()))
		return propResp, nil
	}

	if lastError == nil {
		// all peers were not ready
		return nil, api.ErrNoReadyPeers{MspId: mspId}
	}

	return nil, lastError

}
func (p *peerPool) DeliverClient(mspId string, identity msp.SigningIdentity) (api.DeliverClient, error) {
	poolPeer, err := p.FirstReadyPeer(mspId)
	if err != nil {
		return nil, err
	}
	return poolPeer.DeliverClient(identity)
}

func (p *peerPool) FirstReadyPeer(mspId string) (api.Peer, error) {
	log := p.log.Named(`getFirstReadyPeer`)
	p.storeMx.RLock()
	//check MspId exists
	log.Debug(`Searching peers for MspId`, zap.String(`mspId`, mspId))
	peers, ok := p.store[mspId]
	p.storeMx.RUnlock()

	if !ok {
		log.Error(api.ErrMSPNotFound.Error(), zap.String(`mspId`, mspId))
		return nil, api.ErrMSPNotFound
	}

	//check peers for MspId exists
	if len(peers) == 0 {
		log.Error(api.ErrNoPeersForMSP.Error(), zap.String(`mspId`, mspId))
	}

	log.Debug(`Peers pool`, zap.String(`mspId`, mspId), zap.Int(`peerNum`, len(peers)))

	for _, poolPeer := range peers {
		if poolPeer.ready {
			return poolPeer.peer, nil
		}
	}

	return nil, api.ErrNoReadyPeers{MspId: mspId}
}

func (p *peerPool) Close() error {
	return nil
}

func New(ctx context.Context, log *zap.Logger) api.PeerPool {
	ctx, cancel := context.WithCancel(ctx)

	return &peerPool{
		store:  make(map[string][]*peerPoolPeer),
		log:    log.Named(`PeerPool`),
		ctx:    ctx,
		cancel: cancel,
	}
}
