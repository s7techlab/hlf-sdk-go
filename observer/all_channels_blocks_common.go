package observer

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/common"

	"github.com/s7techlab/hlf-sdk-go/api"
)

type AllChannelBlocksCommon struct {
	*AllChannelsBlocks[*common.Block]

	commonBlocks chan *CommonBlock
	isWork       bool
}

func NewAllChannelBlocksCommon(peerChannels PeerChannelsGetter, blocksDeliver api.BlocksDeliverer, opts ...AllChannelsBlocksOpt) *AllChannelBlocksCommon {
	createStreamWithRetry := CreateBlockStreamWithRetryDelay[*common.Block](DefaultConnectRetryDelay)

	allChsBlocks := NewAllChannelsBlocks[*common.Block](peerChannels, blocksDeliver.Blocks, createStreamWithRetry, opts...)

	return &AllChannelBlocksCommon{AllChannelsBlocks: allChsBlocks}
}

func (a *AllChannelBlocksCommon) Observe(ctx context.Context) <-chan *CommonBlock {
	if a.isWork {
		return a.commonBlocks
	}

	a.commonBlocks = make(chan *CommonBlock)
	go func() {
		a.isWork = true
		defer func() {
			close(a.commonBlocks)
			a.isWork = false
		}()

		blocks := a.AllChannelsBlocks.Observe(ctx)

		for {
			select {
			case <-ctx.Done():
				return

			case cb, ok := <-blocks:
				if !ok {
					return
				}
				if cb == nil {
					continue
				}

				a.commonBlocks <- &CommonBlock{Block: cb}
			}
		}
	}()

	return a.commonBlocks
}
