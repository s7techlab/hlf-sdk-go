package api

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
)

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
