package observer

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type ChannelBlocksWithName[T any] struct {
	Name   string
	Blocks <-chan *Block[T]
}

type ChannelWithChannels[T any] struct {
	channels chan *ChannelBlocksWithName[T]
}

func (cwc *ChannelWithChannels[T]) Observe() chan *ChannelBlocksWithName[T] {
	return cwc.channels
}

func (acb *AllChannelsBlocks[T]) ObserveByChannels(ctx context.Context) *ChannelWithChannels[T] {
	channelWithChannels := &ChannelWithChannels[T]{
		channels: make(chan *ChannelBlocksWithName[T]),
	}

	// ctxObserve using for nested control process without stopped primary context
	ctxObserve, cancel := context.WithCancel(ctx)
	acb.cancelObserve = cancel

	acb.startNotObservedChannelsConcurrently(ctxObserve, acb.initChannelsObservers(), channelWithChannels)

	// init new channels if they are fetched
	go func() {
		defer func() {
			close(channelWithChannels.channels)
		}()

		ticker := time.NewTicker(acb.observePeriod)
		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				acb.startNotObservedChannelsConcurrently(ctxObserve, acb.initChannelsObservers(), channelWithChannels)
			}
		}
	}()

	// closer
	go func() {
		<-ctx.Done()
		acb.Stop()
	}()

	return channelWithChannels
}

func (acb *AllChannelsBlocks[T]) startNotObservedChannelsConcurrently(
	ctx context.Context,
	notObservedChannels []*ChannelBlocks[T],
	channelWithChannels *ChannelWithChannels[T],
) {

	for _, notObservedChannel := range notObservedChannels {
		chBlocks := notObservedChannel

		if _, err := chBlocks.Observe(ctx); err != nil {
			acb.logger.Warn(`init channel observer concurrently`, zap.String("channel", notObservedChannel.channel), zap.Error(err))
		}

		go func() {
			channelWithChannels.channels <- &ChannelBlocksWithName[T]{Name: chBlocks.channel, Blocks: chBlocks.channelWithBlocks}
		}()
	}
}
