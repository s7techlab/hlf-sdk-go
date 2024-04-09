package observer

import (
	"github.com/s7techlab/hlf-sdk-go/api"
	hlfproto "github.com/s7techlab/hlf-sdk-go/block"
)

type AllChannelBlocksParsed struct {
	*AllChannelsBlocks[*hlfproto.Block]
}

func NewAllChannelBlocksParsed(peerChannels PeerChannelsGetter, blocksDeliver api.ParsedBlocksDeliverer, opts ...AllChannelsBlocksOpt) *AllChannelBlocksParsed {
	createStreamWithRetry := CreateBlockStreamWithRetryDelay[*hlfproto.Block](DefaultConnectRetryDelay)

	allChsBlocks := NewAllChannelsBlocks[*hlfproto.Block](peerChannels, blocksDeliver.ParsedBlocks, createStreamWithRetry, opts...)

	return &AllChannelBlocksParsed{AllChannelsBlocks: allChsBlocks}
}
