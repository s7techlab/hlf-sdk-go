package observer

import (
	"github.com/s7techlab/hlf-sdk-go/api"
	hlfproto "github.com/s7techlab/hlf-sdk-go/block"
)

type ChannelsBlocksPeerParsed struct {
	*ChannelsBlocksPeer[*hlfproto.Block]
}

func NewChannelsBlocksPeerParsed(peerChannels PeerChannelsGetter, blocksDeliver api.ParsedBlocksDeliverer, opts ...ChannelsBlocksPeerOpt) *ChannelsBlocksPeerParsed {
	createStreamWithRetry := CreateBlockStreamWithRetryDelay[*hlfproto.Block](DefaultConnectRetryDelay)

	channelsBlocksPeerParsed := NewChannelsBlocksPeer[*hlfproto.Block](peerChannels, blocksDeliver.ParsedBlocks, createStreamWithRetry, opts...)

	return &ChannelsBlocksPeerParsed{ChannelsBlocksPeer: channelsBlocksPeerParsed}
}
