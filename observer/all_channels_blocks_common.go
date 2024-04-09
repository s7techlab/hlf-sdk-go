package observer

import (
	"github.com/hyperledger/fabric-protos-go/common"

	"github.com/s7techlab/hlf-sdk-go/api"
)

type AllChannelBlocksCommon struct {
	*AllChannelsBlocks[*common.Block]
}

func NewAllChannelBlocksCommon(peerChannels PeerChannelsGetter, blocksDeliver api.BlocksDeliverer, opts ...AllChannelsBlocksOpt) *AllChannelBlocksCommon {
	createStreamWithRetry := CreateBlockStreamWithRetryDelay[*common.Block](DefaultConnectRetryDelay)

	allChsBlocks := NewAllChannelsBlocks[*common.Block](peerChannels, blocksDeliver.Blocks, createStreamWithRetry, opts...)

	return &AllChannelBlocksCommon{AllChannelsBlocks: allChsBlocks}
}
