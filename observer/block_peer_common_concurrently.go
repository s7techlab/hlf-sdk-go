package observer

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type ChannelCommonBlocks struct {
	Name   string
	Blocks <-chan *Block
}

type BlocksByChannels struct {
	channels chan *ChannelCommonBlocks
}

func (b *BlocksByChannels) Observe() chan *ChannelCommonBlocks {
	return b.channels
}

func (bp *BlockPeer) ObserveByChannels(ctx context.Context) *BlocksByChannels {
	blocksByChannels := &BlocksByChannels{
		channels: make(chan *ChannelCommonBlocks),
	}

	bp.initChannelsConcurrently(ctx, blocksByChannels)

	// init new channels if they are fetched
	go func() {
		ticker := time.NewTicker(bp.observePeriod)
		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				bp.initChannelsConcurrently(ctx, blocksByChannels)
			}
		}
	}()

	// closer
	go func() {
		<-ctx.Done()
		bp.Stop()
	}()

	return blocksByChannels
}

func (bp *BlockPeer) initChannelsConcurrently(ctx context.Context, blocksByChannels *BlocksByChannels) {
	for channel := range bp.peerChannels.Channels() {
		bp.mu.RLock()
		_, ok := bp.channelObservers[channel]
		bp.mu.RUnlock()

		if !ok {
			bp.logger.Info(`add channel observer concurrently`, zap.String(`channel`, channel))

			blockPeerChannel := bp.peerChannelConcurrently(ctx, channel, blocksByChannels)

			bp.mu.Lock()
			bp.channelObservers[channel] = blockPeerChannel
			bp.mu.Unlock()
		}
	}
}

func (bp *BlockPeer) peerChannelConcurrently(ctx context.Context, channel string, blocksByChannels *BlocksByChannels) *BlockPeerChannel {
	seekFrom := bp.getSeekFrom(channel)

	peerChannel := &BlockPeerChannel{}
	peerChannel.Observer = NewBlockChannel(
		channel,
		bp.blockDeliverer,
		seekFrom,
		WithChannelBlockLogger(bp.logger),
		WithChannelStopRecreateStream(bp.stopRecreateStream))

	_, peerChannel.err = peerChannel.Observer.Observe(ctx)
	if peerChannel.err != nil {
		bp.logger.Warn(`init channel observer`, zap.Error(peerChannel.err))
	}

	blocks := make(chan *Block)
	bp.blocksByChannels[channel] = blocks

	go func() {
		blocksByChannels.channels <- &ChannelCommonBlocks{Name: channel, Blocks: blocks}
	}()

	// channel merger
	go func() {
		for b := range peerChannel.Observer.blocks {
			blocks <- b
		}
	}()

	return peerChannel
}
