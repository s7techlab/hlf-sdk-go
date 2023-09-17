package observer

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type ChannelBlocks struct {
	Name   string
	Blocks <-chan *Block
}

type BlocksByChannels struct {
	channels chan *ChannelBlocks
}

func (b *BlocksByChannels) Observe() chan *ChannelBlocks {
	return b.channels
}

func (bp *BlockPeer) ObserveByChannels(ctx context.Context) (*BlocksByChannels, error) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	blocksByChannels := &BlocksByChannels{
		channels: make(chan *ChannelBlocks),
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

	return blocksByChannels, nil
}

func (bp *BlockPeer) initChannelsConcurrently(ctx context.Context, blocksByChannels *BlocksByChannels) {
	for channel := range bp.peerChannels.Channels() {
		if _, ok := bp.channelObservers[channel]; !ok {
			bp.logger.Info(`add channel observer`, zap.String(`channel`, channel))

			bp.channelObservers[channel] = bp.peerChannelConcurrently(ctx, channel, blocksByChannels)
		}
	}
}

func (bp *BlockPeer) peerChannelConcurrently(ctx context.Context, channel string, blocksByChannels *BlocksByChannels) *blockPeerChannel {
	seekFrom := bp.seekFrom[channel]
	if seekFrom > 0 {
		seekFrom--
	}

	configBlock := bp.configBlocks[channel]

	peerChannel := &blockPeerChannel{}
	peerChannel.observer = NewBlockChannel(
		channel,
		bp.blockDeliverer,
		ChannelSeekFrom(seekFrom),
		WithChannelBlockTransformers(bp.transformers),
		WithChannelConfigBlock(configBlock),
		WithChannelBlockLogger(bp.logger),
		WithChannelStopRecreateStream(bp.stopRecreateStream))

	_, peerChannel.err = peerChannel.observer.Observe(ctx)
	if peerChannel.err != nil {
		bp.logger.Warn(`init channel observer`, zap.Error(peerChannel.err))
	}

	blocks := make(chan *Block)
	bp.blocksByChannels[channel] = blocks

	go func() {
		blocksByChannels.channels <- &ChannelBlocks{Name: channel, Blocks: blocks}
	}()

	// channel merger
	go func() {
		for b := range peerChannel.observer.blocks {
			blocks <- b
		}
	}()

	return peerChannel
}
