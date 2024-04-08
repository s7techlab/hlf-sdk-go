package observer

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/msp"
	"go.uber.org/zap"

	hlfproto "github.com/s7techlab/hlf-sdk-go/block"
)

type (
	ChannelBlocks[T any] struct {
		*Channel

		channelWithBlocks     chan *Block[T]
		blocksDeliverer       func(context.Context, string, msp.SigningIdentity, ...int64) (<-chan T, func() error, error)
		createStreamWithRetry CreateBlockStreamWithRetry[T]

		stopRecreateStream bool

		isWork        bool
		cancelObserve context.CancelFunc
	}

	ChannelBlocksOpts struct {
		*Opts

		// don't recreate stream if it has not any blocks
		stopRecreateStream bool
	}

	ChannelBlocksOpt func(*ChannelBlocksOpts)
)

func WithChannelBlockLogger(logger *zap.Logger) ChannelBlocksOpt {
	return func(opts *ChannelBlocksOpts) {
		opts.Opts.logger = logger
	}
}

func WithChannelStopRecreateStream(stop bool) ChannelBlocksOpt {
	return func(opts *ChannelBlocksOpts) {
		opts.stopRecreateStream = stop
	}
}

var DefaultChannelBlocksOpts = &ChannelBlocksOpts{
	Opts:               DefaultOpts,
	stopRecreateStream: false,
}

func NewChannelBlocks[T any](
	channel string,
	deliverer func(context.Context, string, msp.SigningIdentity, ...int64) (<-chan T, func() error, error),
	createStreamWithRetry CreateBlockStreamWithRetry[T],
	seekFromFetcher SeekFromFetcher,
	opts ...ChannelBlocksOpt,
) *ChannelBlocks[T] {

	channelBlocksOpts := DefaultChannelBlocksOpts
	for _, opt := range opts {
		opt(channelBlocksOpts)
	}

	return &ChannelBlocks[T]{
		Channel: &Channel{
			channel:         channel,
			seekFromFetcher: seekFromFetcher,
			identity:        channelBlocksOpts.identity,
			logger:          channelBlocksOpts.logger.With(zap.String(`channel`, channel)),
		},

		blocksDeliverer:       deliverer,
		createStreamWithRetry: createStreamWithRetry,
		stopRecreateStream:    channelBlocksOpts.stopRecreateStream,
	}
}

func (cb *ChannelBlocks[T]) Stop() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// cb.channelWithBlocks mustn't be closed here, because it is closed elsewhere

	err := cb.Channel.stop()

	// If primary context is done then cancel ctxObserver
	if cb.cancelObserve != nil {
		cb.cancelObserve()
	}

	cb.isWork = false
	return err
}

func (cb *ChannelBlocks[T]) Observe(ctx context.Context) (<-chan *Block[T], error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.isWork {
		return cb.channelWithBlocks, nil
	}

	// ctxObserve using for nested control process without stopped primary context
	ctxObserve, cancel := context.WithCancel(ctx)
	cb.cancelObserve = cancel

	if err := cb.allowToObserve(); err != nil {
		return nil, err
	}

	// Double check
	if err := cb.allowToObserve(); err != nil {
		return nil, err
	}

	cb.channelWithBlocks = make(chan *Block[T])

	go func() {
		cb.isWork = true

		defer close(cb.channelWithBlocks)

		cb.logger.Debug(`creating block stream`)
		incomingBlocks, errCreateStream := cb.createStreamWithRetry(ctxObserve, cb.createStream)
		if errCreateStream != nil {
			return
		}

		cb.logger.Info(`block stream created`)
		for {
			select {
			case incomingBlock, hasMore := <-incomingBlocks:

				var err error
				if !hasMore && !cb.stopRecreateStream {
					cb.logger.Debug(`block stream interrupted, recreate`)
					incomingBlocks, err = cb.createStreamWithRetry(ctx, cb.createStream)
					if err != nil {
						return
					}

					cb.logger.Debug(`block stream recreated`)
					continue
				}

				switch t := any(incomingBlock).(type) {
				case *common.Block:
					if t == nil {
						continue
					}

				case *hlfproto.Block:
					if t == nil {
						continue
					}
				}

				cb.channelWithBlocks <- &Block[T]{
					Channel: cb.channel,
					Block:   incomingBlock,
				}

			case <-ctxObserve.Done():
				if err := cb.Stop(); err != nil {
					cb.lastError = err
				}
				return
			}
		}
	}()

	return cb.channelWithBlocks, nil
}

func (cb *ChannelBlocks[T]) createStream(ctx context.Context) (<-chan T, error) {
	cb.preCreateStream()

	cb.logger.Debug(`connecting to blocks stream, receiving seek offset`,
		zap.Uint64(`attempt`, cb.connectAttempt))

	seekFrom, err := cb.processSeekFrom(ctx)
	if err != nil {
		cb.logger.Warn(`seek from failed`, zap.Error(err))
		return nil, err
	}
	cb.logger.Info(`block seek offset received`, zap.Uint64(`seek from`, seekFrom))

	var (
		blocks <-chan T
		closer func() error
	)
	cb.logger.Debug(`subscribing to blocks stream`)
	blocks, closer, err = cb.blocksDeliverer(ctx, cb.channel, cb.identity, int64(seekFrom))
	if err != nil {
		cb.logger.Warn(`subscribing to blocks stream failed`, zap.Error(err))
		cb.setError(err)
		return nil, fmt.Errorf(`blocks deliverer: %w`, err)
	}
	cb.logger.Info(`subscribed to blocks stream`)

	cb.afterCreateStream(closer)

	// Check close context
	select {
	case <-ctx.Done():
		err = closer()
		return nil, err
	default:
	}

	return blocks, nil
}
