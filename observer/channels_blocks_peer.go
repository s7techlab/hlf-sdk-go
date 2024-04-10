package observer

import (
	"context"
	"sync"
	"time"

	"github.com/hyperledger/fabric/msp"
	"go.uber.org/zap"
)

const DefaultChannelsBLocksPeerRefreshPeriod = 10 * time.Second

type (
	PeerChannelsGetter interface {
		URI() string
		Channels() map[string]*ChannelInfo
	}

	ChannelsBlocksPeer[T any] struct {
		channelObservers map[string]*ChannelBlocks[T]

		blocks chan *Block[T]

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

	ChannelsBlocksPeerOpts struct {
		seekFrom           map[string]uint64
		seekFromFetcher    SeekFromFetcher
		refreshPeriod      time.Duration
		stopRecreateStream bool
		logger             *zap.Logger
	}

	ChannelsBlocksPeerOpt func(*ChannelsBlocksPeerOpts)
)

var DefaultChannelsBlocksPeerOpts = &ChannelsBlocksPeerOpts{
	refreshPeriod: DefaultChannelsBLocksPeerRefreshPeriod,
	logger:        zap.NewNop(),
}

func WithChannelsBlocksPeerLogger(logger *zap.Logger) ChannelsBlocksPeerOpt {
	return func(opts *ChannelsBlocksPeerOpts) {
		opts.logger = logger
	}
}

func WithSeekFrom(seekFrom map[string]uint64) ChannelsBlocksPeerOpt {
	return func(opts *ChannelsBlocksPeerOpts) {
		opts.seekFrom = seekFrom
	}
}

func WithSeekFromFetcher(seekFromFetcher SeekFromFetcher) ChannelsBlocksPeerOpt {
	return func(opts *ChannelsBlocksPeerOpts) {
		opts.seekFromFetcher = seekFromFetcher
	}
}

func WithChannelsBlocksPeerRefreshPeriod(refreshPeriod time.Duration) ChannelsBlocksPeerOpt {
	return func(opts *ChannelsBlocksPeerOpts) {
		if refreshPeriod != 0 {
			opts.refreshPeriod = refreshPeriod
		}
	}
}

func WithBlockStopRecreateStream(stop bool) ChannelsBlocksPeerOpt {
	return func(opts *ChannelsBlocksPeerOpts) {
		opts.stopRecreateStream = stop
	}
}

func NewChannelsBlocksPeer[T any](
	peerChannelsGetter PeerChannelsGetter,
	deliverer func(context.Context, string, msp.SigningIdentity, ...int64) (<-chan T, func() error, error),
	createStreamWithRetry CreateBlockStreamWithRetry[T],
	opts ...ChannelsBlocksPeerOpt,
) *ChannelsBlocksPeer[T] {

	channelsBlocksPeerOpts := DefaultChannelsBlocksPeerOpts
	for _, opt := range opts {
		opt(channelsBlocksPeerOpts)
	}

	return &ChannelsBlocksPeer[T]{
		channelObservers: make(map[string]*ChannelBlocks[T]),
		blocks:           make(chan *Block[T]),

		peerChannelsGetter:    peerChannelsGetter,
		deliverer:             deliverer,
		createStreamWithRetry: createStreamWithRetry,
		observePeriod:         channelsBlocksPeerOpts.refreshPeriod,

		seekFrom:           channelsBlocksPeerOpts.seekFrom,
		seekFromFetcher:    channelsBlocksPeerOpts.seekFromFetcher,
		stopRecreateStream: channelsBlocksPeerOpts.stopRecreateStream,
		logger:             channelsBlocksPeerOpts.logger,
	}
}

func (acb *ChannelsBlocksPeer[T]) Channels() map[string]*Channel {
	acb.mu.RLock()
	defer acb.mu.RUnlock()

	var copyChannels = make(map[string]*Channel, len(acb.channelObservers))
	for key, value := range acb.channelObservers {
		copyChannels[key] = value.Channel
	}

	return copyChannels
}

func (acb *ChannelsBlocksPeer[T]) Stop() {
	// acb.blocks and acb.blocksByChannels mustn't be closed here, because they are closed elsewhere

	acb.mu.RLock()
	for _, c := range acb.channelObservers {
		if err := c.Stop(); err != nil {
			zap.Error(err)
		}
	}
	acb.mu.RUnlock()

	acb.mu.Lock()
	acb.channelObservers = make(map[string]*ChannelBlocks[T])
	acb.mu.Unlock()

	if acb.cancelObserve != nil {
		acb.cancelObserve()
	}

	acb.isWork = false
}

func (acb *ChannelsBlocksPeer[T]) Observe(ctx context.Context) <-chan *Block[T] {
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
		defer ticker.Stop()

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

func (acb *ChannelsBlocksPeer[T]) startNotObservedChannels(ctx context.Context, notObservedChannels []*ChannelBlocks[T]) {
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

func (acb *ChannelsBlocksPeer[T]) initChannelsObservers() []*ChannelBlocks[T] {
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

func (acb *ChannelsBlocksPeer[T]) getSeekFrom(channel string) SeekFromFetcher {
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
