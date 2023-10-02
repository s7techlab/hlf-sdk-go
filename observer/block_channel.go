package observer

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-protos-go/common"
	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/api"
)

type (
	BlockChannel struct {
		*Channel
		blocksDeliverer       api.BlocksDeliverer
		createStreamWithRetry CreateBlockStreamWithRetry
		stopRecreateStream    bool

		blocks chan *common.Block

		isWork        bool
		cancelObserve context.CancelFunc
	}

	BlockChannelOpts struct {
		*Opts

		createStreamWithRetry CreateBlockStreamWithRetry

		// don't recreate stream if it has not any blocks
		stopRecreateStream bool
	}

	BlockChannelOpt func(*BlockChannelOpts)
)

func WithChannelBlockLogger(logger *zap.Logger) BlockChannelOpt {
	return func(opts *BlockChannelOpts) {
		opts.Opts.logger = logger
	}
}

func WithChannelStopRecreateStream(stop bool) BlockChannelOpt {
	return func(opts *BlockChannelOpts) {
		opts.stopRecreateStream = stop
	}
}

var DefaultBlockChannelOpts = &BlockChannelOpts{
	createStreamWithRetry: CreateBlockStreamWithRetryDelay(DefaultConnectRetryDelay),
	Opts:                  DefaultOpts,
}

func NewBlockChannel(channel string, blocksDeliver api.BlocksDeliverer, seekFromFetcher SeekFromFetcher, opts ...BlockChannelOpt) *BlockChannel {
	blockChannelOpts := DefaultBlockChannelOpts
	for _, opt := range opts {
		opt(blockChannelOpts)
	}

	observer := &BlockChannel{
		Channel: &Channel{
			channel:         channel,
			seekFromFetcher: seekFromFetcher,
			identity:        blockChannelOpts.identity,
			logger:          blockChannelOpts.logger.With(zap.String(`channel`, channel)),
		},

		blocksDeliverer:       blocksDeliver,
		createStreamWithRetry: blockChannelOpts.createStreamWithRetry,
		stopRecreateStream:    blockChannelOpts.stopRecreateStream,
	}

	return observer
}

func (c *BlockChannel) Observe(ctx context.Context) (<-chan *common.Block, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.isWork {
		return c.blocks, nil
	}

	// ctxObserve using for nested control process without stopped primary context
	ctxObserve, cancel := context.WithCancel(ctx)
	c.cancelObserve = cancel

	if err := c.allowToObserve(); err != nil {
		return nil, err
	}

	// Double check
	if err := c.allowToObserve(); err != nil {
		return nil, err
	}

	c.blocks = make(chan *common.Block)

	go func() {
		c.isWork = true

		c.logger.Debug(`creating block stream`)
		incomingBlocks, errCreateStream := c.createStreamWithRetry(ctxObserve, c.createStream)
		if errCreateStream != nil {
			return
		}

		c.logger.Info(`block stream created`)
		for {
			select {
			case incomingBlock, hasMore := <-incomingBlocks:

				var err error
				if !hasMore && !c.stopRecreateStream {
					c.logger.Debug(`block stream interrupted, recreate`)
					incomingBlocks, err = c.createStreamWithRetry(ctx, c.createStream)
					if err != nil {
						return
					}

					c.logger.Debug(`block stream recreated`)
					continue
				}

				if incomingBlock == nil {
					continue
				}

				c.blocks <- incomingBlock

			case <-ctxObserve.Done():
				if err := c.Stop(); err != nil {
					c.lastError = err
				}
				return
			}
		}
	}()

	return c.blocks, nil
}

func (c *BlockChannel) Stop() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	err := c.Channel.stop()

	// If primary context is done then cancel ctxObserver
	if c.cancelObserve != nil {
		c.cancelObserve()
	}

	c.isWork = false
	return err
}

func (c *BlockChannel) createStream(ctx context.Context) (<-chan *common.Block, error) {
	c.preCreateStream()

	c.logger.Debug(`connecting to blocks stream, receiving seek offset`,
		zap.Uint64(`attempt`, c.connectAttempt))

	seekFrom, err := c.processSeekFrom(ctx)
	if err != nil {
		c.logger.Warn(`seek from failed`, zap.Error(err))
		return nil, err
	}
	c.logger.Info(`block seek offset received`, zap.Uint64(`seek from`, seekFrom))

	var (
		blocks <-chan *common.Block
		closer func() error
	)
	c.logger.Debug(`subscribing to blocks stream`)
	blocks, closer, err = c.blocksDeliverer.Blocks(ctx, c.channel, c.identity, int64(seekFrom))
	if err != nil {
		c.logger.Warn(`subscribing to blocks stream failed`, zap.Error(err))
		c.setError(err)
		return nil, fmt.Errorf(`blocks deliverer: %w`, err)
	}
	c.logger.Info(`subscribed to blocks stream`)

	c.afterCreateStream(closer)

	// Check close context
	select {
	case <-ctx.Done():
		err = closer()
		return nil, err
	default:
	}

	return blocks, nil
}
