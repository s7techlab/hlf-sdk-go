package api

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// CSCC describes Configuration System Chaincode (CSCC)
type CSCC interface {
	// JoinChain allows to join channel using presented genesis block
	JoinChain(ctx context.Context, channelName string, genesisBlock *common.Block) error
	// GetConfigBlock returns genesis block of channel
	GetConfigBlock(ctx context.Context, channelName string) (*common.Block, error)
	// GetChannelConfig returns channel configuration
	GetChannelConfig(ctx context.Context, channelName string) (*common.Config, error)
	// GetChannels returns list of joined channels
	GetChannels(ctx context.Context) (*peer.ChannelQueryResponse, error)
}
