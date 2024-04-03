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
	for channel := range pbp.blockPeer.peerChannels.Channels() {
		pbp.mu.RLock()
		_, ok := pbp.parsedChannelObservers[channel]
		pbp.mu.RUnlock()

		if !ok {
			pbp.blockPeer.logger.Info(`add parsed channel observer concurrently`, zap.String(`channel`, channel))

			parsedBlockPeerChannel := pbp.peerParsedChannelConcurrently(ctx, channel, blocksByChannels)

			pbp.mu.Lock()
			pbp.parsedChannelObservers[channel] = parsedBlockPeerChannel
			pbp.mu.Unlock()
		}
	}
}

func (pbp *ParsedBlockPeer) peerParsedChannelConcurrently(ctx context.Context, channel string, blocksByChannels *ParsedBlocksByChannels) *ParsedBlockPeerChannel {
	seekFrom := pbp.blockPeer.getSeekFrom(channel)

	commonBlockChannel := NewBlockChannel(
		channel,
		pbp.blockPeer.blockDeliverer,
		seekFrom,
		WithChannelBlockLogger(pbp.blockPeer.logger),
		WithChannelStopRecreateStream(pbp.blockPeer.stopRecreateStream))

	configBlock := pbp.configBlocks[channel]

	peerParsedChannel := &ParsedBlockPeerChannel{}
	peerParsedChannel.Observer = NewParsedBlockChannel(
		commonBlockChannel,
		WithParsedChannelBlockTransformers(pbp.transformers),
		WithParsedChannelConfigBlock(configBlock))

	_, peerParsedChannel.err = peerParsedChannel.Observer.Observe(ctx)
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
		for b := range peerParsedChannel.Observer.parsedBlocks {
			blocks <- b
		}
	}()

	return peerParsedChannel
}
