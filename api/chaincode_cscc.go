// +build !fabric2

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
	// GetConfigTree returns configuration tree of channel
	GetConfigTree(ctx context.Context, channelName string) (*peer.ConfigTree, error)
	// Channels returns list of joined channels
	Channels(ctx context.Context) (*peer.ChannelQueryResponse, error)
}
