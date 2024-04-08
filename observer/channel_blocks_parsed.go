package observer

import (
	"github.com/s7techlab/hlf-sdk-go/api"
	hlfproto "github.com/s7techlab/hlf-sdk-go/block"
)

type (
	ChannelBlocksParsed struct {
		*ChannelBlocks[*hlfproto.Block]
	}
)

func NewChannelBlocksParsed(channel string, blocksDeliver api.ParsedBlocksDeliverer, seekFromFetcher SeekFromFetcher, opts ...ChannelBlocksOpt) *ChannelBlocksParsed {
	createStreamWithRetry := CreateBlockStreamWithRetryDelay[*hlfproto.Block](DefaultConnectRetryDelay)

	chBlocks := NewChannelBlocks[*hlfproto.Block](channel, blocksDeliver.ParsedBlocks, createStreamWithRetry, seekFromFetcher, opts...)

	return &ChannelBlocksParsed{ChannelBlocks: chBlocks}
}
