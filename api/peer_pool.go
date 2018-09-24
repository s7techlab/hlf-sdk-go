package api

import (
	"context"
	"time"

	"github.com/hyperledger/fabric/protos/peer"
	"google.golang.org/grpc/connectivity"
)

const (
	ErrPeerAlreadySet     = Error(`peer already set`)
	ErrNoPeersForMSP      = Error(`no peers for presented MSP`)
	ErrNoReadyPeersForMSP = Error(`no ready peers for presented MSP`)
	ErrMSPNotFound        = Error(`MSP not found`)
)

type PeerPool interface {
	Add(mspId string, peer Peer, strategy PeerPoolCheckStrategy) error
	Process(mspId string, context context.Context, proposal *peer.SignedProposal) (*peer.ProposalResponse, error)
	//DeliverClient(mspId string) DeliverClient
	Close() error
}

type PeerPoolCheckStrategy func(peer Peer, alive chan bool, ctx context.Context)

func StrategyGRPC(d time.Duration) PeerPoolCheckStrategy {
	return func(peer Peer, alive chan bool, ctx context.Context) {
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
