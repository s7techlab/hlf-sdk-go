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

func (pbp *ParsedBlockPeer) ObserveByChannels(ctx context.Context) (*ParsedBlocksByChannels, error) {
	pbp.mu.Lock()
	defer pbp.mu.Unlock()

	// need to start default block peer observing
	_, _ = pbp.blockPeer.ObserveByChannels(ctx)

	blocksByChannels := &ParsedBlocksByChannels{
		channels: make(chan *ChannelParsedBlocks),
	}

	pbp.initParsedChannelsConcurrently(ctx, blocksByChannels)

	// init new channels if they are fetched
	go func() {
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

	return blocksByChannels, nil
}

func (pbp *ParsedBlockPeer) initParsedChannelsConcurrently(ctx context.Context, blocksByChannels *ParsedBlocksByChannels) {
	for _, commonBlockChannelObserver := range pbp.blockPeer.ChannelObservers() {
		channel := commonBlockChannelObserver.observer.channel
		if _, ok := pbp.parsedChannelObservers[channel]; !ok {
			pbp.blockPeer.logger.Info(`add parsed channel observer concurrently`, zap.String(`channel`, channel))

			pbp.parsedChannelObservers[channel] = pbp.peerParsedChannelConcurrently(ctx, channel, blocksByChannels, commonBlockChannelObserver.observer)
		}
	}
}

func (pbp *ParsedBlockPeer) peerParsedChannelConcurrently(ctx context.Context, channel string, blocksByChannels *ParsedBlocksByChannels, commonBlockChannel *BlockChannel) *parsedBlockPeerChannel {
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
