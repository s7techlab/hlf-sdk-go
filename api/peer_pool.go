package api

import (
	"context"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
	"google.golang.org/grpc/connectivity"
)

const (
	ErrNoPeersForMSP = Error(`no peers for MSP`)
	ErrMSPNotFound   = Error(`MSP not found`)
	ErrPeerNotReady  = Error(`peer not ready`)

	DefaultGrpcCheckPeriod = 5 * time.Second
)

type ErrNoReadyPeers struct {
	MspId string
}

func (e ErrNoReadyPeers) Error() string {
	return fmt.Sprintf("no ready peers for MspId: %s", e.MspId)
}

type MSPPeerPool interface {
	Peers() []Peer
	FirstReadyPeer() (Peer, error)
	Process(ctx context.Context, proposal *peer.SignedProposal) (*peer.ProposalResponse, error)
	DeliverClient(identity msp.SigningIdentity) (DeliverClient, error)
}

type PeerPool interface {
	GetPeers() map[string][]Peer
	GetMSPPeers(mspID string) []Peer
	FirstReadyPeer(mspID string) (Peer, error)
	Add(mspId string, peer Peer, strategy PeerPoolCheckStrategy) error
	EndorseOnMSP(ctx context.Context, mspId string, proposal *peer.SignedProposal) (*peer.ProposalResponse, error)
	EndorseOnMSPs(ctx context.Context, endorsingMspIDs []string, proposal *peer.SignedProposal) ([]*peer.ProposalResponse, error)
	DeliverClient(mspId string, identity msp.SigningIdentity) (DeliverClient, error)
	Close() error
}

type PeerPoolCheckStrategy func(ctx context.Context, peer Peer, alive chan bool)

func StrategyGRPC(d time.Duration) PeerPoolCheckStrategy {
	return func(ctx context.Context, peer Peer, alive chan bool) {
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
