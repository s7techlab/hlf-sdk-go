package api

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
	"google.golang.org/grpc"
)

type Endorser interface {
	// Endorse sends proposal to endorsing peer and returns its result
	Endorse(ctx context.Context, proposal *peer.SignedProposal) (*peer.ProposalResponse, error)
}

// Peer is common interface for endorsing peer
type Peer interface {
	Querier

	Endorser

	// DeliverClient returns DeliverClient
	DeliverClient(identity msp.SigningIdentity) (DeliverClient, error)
	// Uri returns url used for grpc connection
	Uri() string
	// Conn returns instance of grpc connection
	Conn() *grpc.ClientConn
	// Close terminates peer connection
	Close() error
}

// PeerProcessor is interface for processing transaction
type PeerProcessor interface {
	// CreateProposal creates signed proposal for presented cc, function and args using signing identity
	CreateProposal(chaincodeName string, identity msp.SigningIdentity, fn string, args [][]byte, transArgs TransArgs) (*peer.SignedProposal, ChaincodeTx, error)
	// Send sends signed proposal to endorsing peers and collects their responses
	Send(ctx context.Context, proposal *peer.SignedProposal, endorsingMspIDs []string, pool PeerPool) ([]*peer.ProposalResponse, error)
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
