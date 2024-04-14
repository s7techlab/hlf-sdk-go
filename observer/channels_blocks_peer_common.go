package observer

import (
	"github.com/hyperledger/fabric-protos-go/common"

	"github.com/s7techlab/hlf-sdk-go/api"
)

type ChannelsBlocksPeerCommon struct {
	*ChannelsBlocksPeer[*common.Block]
}

func NewChannelsBlocksPeerCommon(peerChannels PeerChannelsGetter, blocksDeliver api.BlocksDeliverer, opts ...ChannelsBlocksPeerOpt) *ChannelsBlocksPeerCommon {
	createStreamWithRetry := CreateBlockStreamWithRetryDelay[*common.Block](DefaultConnectRetryDelay)

	channelsBlocksPeerCommon := NewChannelsBlocksPeer[*common.Block](peerChannels, blocksDeliver.Blocks, createStreamWithRetry, opts...)

	return &ChannelsBlocksPeerCommon{ChannelsBlocksPeer: channelsBlocksPeerCommon}
}
