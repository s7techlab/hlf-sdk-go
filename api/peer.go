package api

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
	"google.golang.org/grpc"
)

type Endorser interface {
	// Endorse sends proposal to endorsing peer and returns its result
	Endorse(ctx context.Context, proposal *peer.SignedProposal) (*peer.ProposalResponse, error)
}

type ChannelListGetter interface {
	GetChannels(ctx context.Context) (*peer.ChannelQueryResponse, error)
}

type ChainInfoGetter interface {
	GetChainInfo(ctx context.Context, channel string) (*common.BlockchainInfo, error)
}

// Peer is common interface for endorsing peer
type Peer interface {
	Querier

	Endorser

	ChannelListGetter

	ChainInfoGetter

	BlocksDeliverer

	EventsDeliverer

	// DeliverClient returns DeliverClient
	DeliverClient(identity msp.SigningIdentity) (DeliverClient, error)
	// Uri returns url used for grpc connection
	Uri() string
	// Conn returns instance of grpc connection
	Conn() *grpc.ClientConn
	// Close terminates peer connection
	Close() error
}

// PeerEndorseError describes peer endorse error
// TODO currently not working cause peer embeds error in string
type PeerEndorseError struct {
	Status  int32
	Message string
}

func (e PeerEndorseError) Error() string {
	return fmt.Sprintf("failed to endorse: %s (code: %d)", e.Message, e.Status)
}
