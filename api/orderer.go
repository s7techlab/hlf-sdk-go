package api

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric/msp"
)

type Orderer interface {
	// Broadcast sends envelope to orderer and returns it's result
	Broadcast(ctx context.Context, envelope *common.Envelope) (*orderer.BroadcastResponse, error)
	// Deliver fetches block from orderer by envelope
	Deliver(ctx context.Context, envelope *common.Envelope) (*common.Block, error)
	// GetConfigBlock returns last config block
	GetConfigBlock(ctx context.Context, signer msp.SigningIdentity, channelName string) (*common.Block, error)
}
