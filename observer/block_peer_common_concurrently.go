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
	bp.mu.Lock()
	defer bp.mu.Unlock()

	for channel := range bp.peerChannels.Channels() {
		if _, ok := bp.channelObservers[channel]; !ok {
			bp.logger.Info(`add channel observer concurrently`, zap.String(`channel`, channel))

			bp.channelObservers[channel] = bp.peerChannelConcurrently(ctx, channel, blocksByChannels)
		}
	}
}

func (bp *BlockPeer) peerChannelConcurrently(ctx context.Context, channel string, blocksByChannels *BlocksByChannels) *blockPeerChannel {
	seekFrom := bp.seekFrom[channel]
	if seekFrom > 0 {
		// it must be -1, because start position here is excluded from array
		// https://github.com/s7techlab/hlf-sdk-go/blob/master/proto/seek.go#L15
		seekFrom--
	}

	peerChannel := &blockPeerChannel{}
	peerChannel.observer = NewBlockChannel(
		channel,
		bp.blockDeliverer,
		ChannelSeekFrom(seekFrom),
		WithChannelBlockLogger(bp.logger),
		WithChannelStopRecreateStream(bp.stopRecreateStream))

	_, peerChannel.err = peerChannel.observer.Observe(ctx)
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
		for b := range peerChannel.observer.blocks {
			blocks <- b
		}
	}()

	return peerChannel
}
