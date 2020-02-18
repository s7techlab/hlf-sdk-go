package api

import (
	"context"
	"fmt"
	"time"

	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/peer"
	"google.golang.org/grpc/connectivity"
)

const (
	ErrPeerAlreadySet = Error(`peer already set`)
	ErrNoPeersForMSP  = Error(`no peers for presented MSP`)
	//ErrNoReadyPeersForMSP = Error(`no ready peers for presented MSP`)
	ErrMSPNotFound  = Error(`MSP not found`)
	ErrPeerNotReady = Error(`peer not ready`)
)

type ErrNoReadyPeers struct {
	MspId string
}

func (e ErrNoReadyPeers) Error() string {
	return fmt.Sprintf("no ready peers for MspId: %s", e.MspId)
}

type PeerPool interface {
	Add(mspId string, peer Peer, strategy PeerPoolCheckStrategy) error
	Process(ctx context.Context, mspId string, proposal *peer.SignedProposal) (*peer.ProposalResponse, error)
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
