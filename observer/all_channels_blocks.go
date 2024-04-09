package observer

import (
	"context"
	"sync"
	"time"

	"github.com/hyperledger/fabric/msp"
	"go.uber.org/zap"
)

const DefaultAllChannelsBlocksObservePeriod = 10 * time.Second

type (
	PeerChannelsGetter interface {
		Uri() string
		Channels() map[string]*ChannelInfo
	}

	AllChannelsBlocks[T any] struct {
		channelObservers map[string]*ChannelBlocks[T]

		blocks           chan *Block[T]
		blocksByChannels map[string]chan *Block[T]

		peerChannelsGetter    PeerChannelsGetter
		deliverer             func(context.Context, string, msp.SigningIdentity, ...int64) (<-chan T, func() error, error)
		createStreamWithRetry CreateBlockStreamWithRetry[T]

		observePeriod time.Duration

		// seekFrom has a higher priority than seekFromFetcher (look getSeekFrom method)
		seekFrom           map[string]uint64
		seekFromFetcher    SeekFromFetcher
		stopRecreateStream bool

		isWork        bool
		cancelObserve context.CancelFunc

		mu     sync.RWMutex
		logger *zap.Logger
	}

	AllChannelsBlocksOpts struct {
		seekFrom           map[string]uint64
		seekFromFetcher    SeekFromFetcher
		observePeriod      time.Duration
		stopRecreateStream bool
		logger             *zap.Logger
	}

	AllChannelsBlocksOpt func(*AllChannelsBlocksOpts)
)

var DefaultAllChannelsBlocksOpts = &AllChannelsBlocksOpts{
	observePeriod: DefaultAllChannelsBlocksObservePeriod,
	logger:        zap.NewNop(),
}

func WithAllChannelsBlocksLogger(logger *zap.Logger) AllChannelsBlocksOpt {
	return func(opts *AllChannelsBlocksOpts) {
		opts.logger = logger
	}
}

func WithSeekFrom(seekFrom map[string]uint64) AllChannelsBlocksOpt {
	return func(opts *AllChannelsBlocksOpts) {
		opts.seekFrom = seekFrom
	}
}

func WithSeekFromFetcher(seekFromFetcher SeekFromFetcher) AllChannelsBlocksOpt {
	return func(opts *AllChannelsBlocksOpts) {
		opts.seekFromFetcher = seekFromFetcher
	}
}

func WithAllChannelsBlocksObservePeriod(observePeriod time.Duration) AllChannelsBlocksOpt {
	return func(opts *AllChannelsBlocksOpts) {
		if observePeriod != 0 {
			opts.observePeriod = observePeriod
		}
	}
}

func WithBlockStopRecreateStream(stop bool) AllChannelsBlocksOpt {
	return func(opts *AllChannelsBlocksOpts) {
		opts.stopRecreateStream = stop
	}
}

func NewAllChannelsBlocks[T any](
	peerChannelsGetter PeerChannelsGetter,
	deliverer func(context.Context, string, msp.SigningIdentity, ...int64) (<-chan T, func() error, error),
	createStreamWithRetry CreateBlockStreamWithRetry[T],
	opts ...AllChannelsBlocksOpt,
) *AllChannelsBlocks[T] {

	allChannelsBlocksOpts := DefaultAllChannelsBlocksOpts
	for _, opt := range opts {
		opt(allChannelsBlocksOpts)
	}

	return &AllChannelsBlocks[T]{
		channelObservers: make(map[string]*ChannelBlocks[T]),
		blocks:           make(chan *Block[T]),
		blocksByChannels: make(map[string]chan *Block[T]),

		peerChannelsGetter:    peerChannelsGetter,
		deliverer:             deliverer,
		createStreamWithRetry: createStreamWithRetry,
		observePeriod:         allChannelsBlocksOpts.observePeriod,

		seekFrom:           allChannelsBlocksOpts.seekFrom,
		seekFromFetcher:    allChannelsBlocksOpts.seekFromFetcher,
		stopRecreateStream: allChannelsBlocksOpts.stopRecreateStream,
		logger:             allChannelsBlocksOpts.logger,
	}
}

func (acb *AllChannelsBlocks[T]) Channels() map[string]*Channel {
	acb.mu.RLock()
	defer acb.mu.RUnlock()

	var copyChannels = make(map[string]*Channel, len(acb.channelObservers))
	for key, value := range acb.channelObservers {
		copyChannels[key] = value.Channel
	}

	return copyChannels
}

func (acb *AllChannelsBlocks[T]) Stop() {
	acb.mu.Lock()
	defer acb.mu.Unlock()

	// acb.blocks and acb.blocksByChannels mustn't be closed here, because they are closed elsewhere

	for _, c := range acb.channelObservers {
		if err := c.Stop(); err != nil {
			zap.Error(err)
		}
	}

	acb.channelObservers = make(map[string]*ChannelBlocks[T])

	if acb.cancelObserve != nil {
		acb.cancelObserve()
	}

	acb.isWork = false
}

func (acb *AllChannelsBlocks[T]) Observe(ctx context.Context) <-chan *Block[T] {
	if acb.isWork {
		return acb.blocks
	}

	// ctxObserve using for nested control process without stopped primary context
	ctxObserve, cancel := context.WithCancel(ctx)
	acb.cancelObserve = cancel

	acb.startNotObservedChannels(ctxObserve, acb.initChannelsObservers())

	acb.blocks = make(chan *Block[T])

	// init new channels if they are fetched
	go func() {
		acb.isWork = true
		defer close(acb.blocks)

		ticker := time.NewTicker(acb.observePeriod)
		for {
			select {
			case <-ctxObserve.Done():
				acb.Stop()
				return

			case <-ticker.C:
				acb.startNotObservedChannels(ctxObserve, acb.initChannelsObservers())
			}
		}
	}()

	return acb.blocks
}

func (acb *AllChannelsBlocks[T]) startNotObservedChannels(ctx context.Context, notObservedChannels []*ChannelBlocks[T]) {
	for _, notObservedChannel := range notObservedChannels {
		chBlocks := notObservedChannel

		if _, err := chBlocks.Observe(ctx); err != nil {
			acb.logger.Warn(`init channel observer`, zap.String("channel", notObservedChannel.channel), zap.Error(err))
		}

		// channel merger
		go func() {
			for b := range chBlocks.channelWithBlocks {
				acb.blocks <- b
			}
		}()
	}
}

func (acb *AllChannelsBlocks[T]) initChannelsObservers() []*ChannelBlocks[T] {
	var notObservedChannels []*ChannelBlocks[T]

	for channel := range acb.peerChannelsGetter.Channels() {
		acb.mu.RLock()
		_, ok := acb.channelObservers[channel]
		acb.mu.RUnlock()

		if !ok {
			acb.logger.Info(`add channel observer`, zap.String(`channel`, channel))

			seekFrom := acb.getSeekFrom(channel)

			chBlocks := NewChannelBlocks[T](
				channel,
				acb.deliverer,
				acb.createStreamWithRetry,
				seekFrom,
				WithChannelBlockLogger(acb.logger),
				WithChannelStopRecreateStream(acb.stopRecreateStream))

			acb.mu.Lock()
			acb.channelObservers[channel] = chBlocks
			acb.mu.Unlock()

			notObservedChannels = append(notObservedChannels, chBlocks)
		}
	}

	return notObservedChannels
}

func (acb *AllChannelsBlocks[T]) getSeekFrom(channel string) SeekFromFetcher {
	seekFrom := ChannelSeekOldest()
	// at first check seekFrom var, if it is empty, check seekFromFetcher
	acb.mu.RLock()
	seekFromNum, exist := acb.seekFrom[channel]
	acb.mu.RUnlock()
	if exist {
		seekFrom = ChannelSeekFrom(seekFromNum - 1)
	} else {
		// if seekFromFetcher is also empty, use ChannelSeekOldest
		if acb.seekFromFetcher != nil {
			seekFrom = acb.seekFromFetcher
		}
	}

	return seekFrom
}
