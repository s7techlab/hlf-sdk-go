package observer

import (
	"context"

	"github.com/s7techlab/hlf-sdk-go/api"
	hlfproto "github.com/s7techlab/hlf-sdk-go/block"
)

type AllChannelBlocksParsed struct {
	*AllChannelsBlocks[*hlfproto.Block]

	parsedBlocks chan *ParsedBlock
	isWork       bool
}

func NewAllChannelBlocksParsed(peerChannels PeerChannelsGetter, blocksDeliver api.ParsedBlocksDeliverer, opts ...AllChannelsBlocksOpt) *AllChannelBlocksParsed {
	createStreamWithRetry := CreateBlockStreamWithRetryDelay[*hlfproto.Block](DefaultConnectRetryDelay)

	allChsBlocks := NewAllChannelsBlocks[*hlfproto.Block](peerChannels, blocksDeliver.ParsedBlocks, createStreamWithRetry, opts...)

	return &AllChannelBlocksParsed{AllChannelsBlocks: allChsBlocks}
}

func (a *AllChannelBlocksParsed) Observe(ctx context.Context) <-chan *ParsedBlock {
	if a.isWork {
		return a.parsedBlocks
	}

	a.parsedBlocks = make(chan *ParsedBlock)
	go func() {
		a.isWork = true
		defer func() {
			close(a.parsedBlocks)
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

				a.parsedBlocks <- &ParsedBlock{Block: cb}
			}
		}
	}()

	return a.parsedBlocks
}
