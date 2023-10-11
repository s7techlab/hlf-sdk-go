package observer

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type ChannelParsedBlocks struct {
	Name   string
	Blocks <-chan *ParsedBlock
}

type ParsedBlocksByChannels struct {
	channels chan *ChannelParsedBlocks
}

func (p *ParsedBlocksByChannels) Observe() chan *ChannelParsedBlocks {
	return p.channels
}

func (pbp *ParsedBlockPeer) ObserveByChannels(ctx context.Context) *ParsedBlocksByChannels {
	blocksByChannels := &ParsedBlocksByChannels{
		channels: make(chan *ChannelParsedBlocks),
	}

	pbp.initParsedChannelsConcurrently(ctx, blocksByChannels)

	// init new channels if they are fetched
	go func() {
		pbp.isWork = true

		ticker := time.NewTicker(pbp.blockPeer.observePeriod)
		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				pbp.initParsedChannelsConcurrently(ctx, blocksByChannels)
			}
		}
	}()

	// closer
	go func() {
		<-ctx.Done()
		pbp.Stop()
	}()

	return blocksByChannels
}

func (pbp *ParsedBlockPeer) initParsedChannelsConcurrently(ctx context.Context, blocksByChannels *ParsedBlocksByChannels) {
	pbp.mu.Lock()
	defer pbp.mu.Unlock()

	for channel := range pbp.blockPeer.peerChannels.Channels() {
		if _, ok := pbp.parsedChannelObservers[channel]; !ok {
			pbp.blockPeer.logger.Info(`add parsed channel observer concurrently`, zap.String(`channel`, channel))

			pbp.parsedChannelObservers[channel] = pbp.peerParsedChannelConcurrently(ctx, channel, blocksByChannels)
		}
	}
}

func (pbp *ParsedBlockPeer) peerParsedChannelConcurrently(ctx context.Context, channel string, blocksByChannels *ParsedBlocksByChannels) *parsedBlockPeerChannel {
	seekFrom := pbp.blockPeer.seekFrom[channel]
	if seekFrom > 0 {
		seekFrom--
	}

	commonBlockChannel := NewBlockChannel(channel, pbp.blockPeer.blockDeliverer, ChannelSeekFrom(seekFrom),
		WithChannelBlockLogger(pbp.blockPeer.logger), WithChannelStopRecreateStream(pbp.blockPeer.stopRecreateStream))

	configBlock := pbp.configBlocks[channel]

	peerParsedChannel := &parsedBlockPeerChannel{}
	peerParsedChannel.observer = NewParsedBlockChannel(
		commonBlockChannel,
		WithParsedChannelBlockTransformers(pbp.transformers),
		WithParsedChannelConfigBlock(configBlock))

	_, peerParsedChannel.err = peerParsedChannel.observer.Observe(ctx)
	if peerParsedChannel.err != nil {
		pbp.blockPeer.logger.Warn(`init parsed channel observer`, zap.Error(peerParsedChannel.err))
	}

	blocks := make(chan *ParsedBlock)
	pbp.blocksByChannels[channel] = blocks

	go func() {
		blocksByChannels.channels <- &ChannelParsedBlocks{Name: channel, Blocks: blocks}
	}()

	// channel merger
	go func() {
		for b := range peerParsedChannel.observer.blocks {
			blocks <- b
		}
	}()

	return peerParsedChannel
}
