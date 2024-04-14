package observer

import (
	"github.com/hyperledger/fabric-protos-go/common"

	"github.com/s7techlab/hlf-sdk-go/api"
)

type (
	ChannelBlocksCommon struct {
		*ChannelBlocks[*common.Block]
	}
)

func NewChannelBlocksCommon(channel string, blocksDeliver api.BlocksDeliverer, seekFromFetcher SeekFromFetcher, opts ...ChannelBlocksOpt) *ChannelBlocksCommon {
	createStreamWithRetry := CreateBlockStreamWithRetryDelay[*common.Block](DefaultConnectRetryDelay)

	chBlocks := NewChannelBlocks[*common.Block](channel, blocksDeliver.Blocks, createStreamWithRetry, seekFromFetcher, opts...)

	return &ChannelBlocksCommon{ChannelBlocks: chBlocks}
}
