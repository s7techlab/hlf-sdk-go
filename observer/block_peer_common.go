package observer

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/api"
)

const DefaultBlockPeerObservePeriod = 10 * time.Second

type (
	BlockPeer struct {
		mu sync.RWMutex

		peerChannels       PeerChannels
		blockDeliverer     api.BlocksDeliverer
		channelObservers   map[string]*BlockPeerChannel
		seekFrom           map[string]uint64
		observePeriod      time.Duration
		stopRecreateStream bool
		logger             *zap.Logger

		blocks           chan *Block
		blocksByChannels map[string]chan *Block

		isWork        bool
		cancelObserve context.CancelFunc
	}

	BlockPeerChannel struct {
		Observer *BlockChannel
		err      error
	}

	BlockPeerOpts struct {
		seekFrom           map[string]uint64
		observePeriod      time.Duration
		stopRecreateStream bool
		logger             *zap.Logger
	}

	BlockPeerOpt func(*BlockPeerOpts)

	ChannelStatus struct {
		Status ChannelObserverStatus
		Err    error
	}
)

var DefaultBlockPeerOpts = &BlockPeerOpts{
	observePeriod: DefaultBlockPeerObservePeriod,
	logger:        zap.NewNop(),
}

func WithBlockPeerLogger(logger *zap.Logger) BlockPeerOpt {
	return func(opts *BlockPeerOpts) {
		opts.logger = logger
	}
}

func WithSeekFrom(seekFrom map[string]uint64) BlockPeerOpt {
	return func(opts *BlockPeerOpts) {
		opts.seekFrom = seekFrom
	}
}

func WithBlockPeerObservePeriod(observePeriod time.Duration) BlockPeerOpt {
	return func(opts *BlockPeerOpts) {
		if observePeriod != 0 {
			opts.observePeriod = observePeriod
		}
	}
}

func WithBlockStopRecreateStream(stop bool) BlockPeerOpt {
	return func(opts *BlockPeerOpts) {
		opts.stopRecreateStream = stop
	}
}

func NewBlockPeer(peerChannels PeerChannels, blockDeliverer api.BlocksDeliverer, opts ...BlockPeerOpt) *BlockPeer {
	blockPeerOpts := DefaultBlockPeerOpts
	for _, opt := range opts {
		opt(blockPeerOpts)
	}

	blockPeer := &BlockPeer{
		peerChannels:       peerChannels,
		blockDeliverer:     blockDeliverer,
		channelObservers:   make(map[string]*BlockPeerChannel),
		blocks:             make(chan *Block),
		blocksByChannels:   make(map[string]chan *Block),
		seekFrom:           blockPeerOpts.seekFrom,
		observePeriod:      blockPeerOpts.observePeriod,
		stopRecreateStream: blockPeerOpts.stopRecreateStream,
		logger:             blockPeerOpts.logger,
	}

	return blockPeer
}

func (bp *BlockPeer) ChannelObservers() map[string]*BlockPeerChannel {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	var copyChannelObservers = make(map[string]*BlockPeerChannel, len(bp.channelObservers))
	for key, value := range bp.channelObservers {
		copyChannelObservers[key] = value
	}

	return copyChannelObservers
}

func (bp *BlockPeer) Observe(ctx context.Context) <-chan *Block {
	if bp.isWork {
		return bp.blocks
	}

	// ctxObserve using for nested control process without stopped primary context
	ctxObserve, cancel := context.WithCancel(ctx)
	bp.cancelObserve = cancel

	bp.initChannels(ctxObserve)

	// init new channels if they are fetched
	go func() {
		bp.isWork = true

		ticker := time.NewTicker(bp.observePeriod)
		for {
			select {
			case <-ctxObserve.Done():
				bp.Stop()
				return

			case <-ticker.C:
				bp.initChannels(ctxObserve)
			}
		}
	}()

	return bp.blocks
}

func (bp *BlockPeer) Stop() {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	// bp.blocks and bp.blocksByChannels mustn't be closed here, because they are closed elsewhere

	for _, c := range bp.channelObservers {
		if err := c.Observer.Stop(); err != nil {
			zap.Error(err)
		}
	}

	bp.channelObservers = make(map[string]*BlockPeerChannel)

	if bp.cancelObserve != nil {
		bp.cancelObserve()
	}

	bp.isWork = false
}

func (bp *BlockPeer) initChannels(ctx context.Context) {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	for channel := range bp.peerChannels.Channels() {
		if _, ok := bp.channelObservers[channel]; !ok {
			bp.logger.Info(`add channel Observer`, zap.String(`channel`, channel))

			bp.channelObservers[channel] = bp.peerChannel(ctx, channel)
		}
	}
}

func (bp *BlockPeer) peerChannel(ctx context.Context, channel string) *BlockPeerChannel {
	seekFrom := bp.seekFrom[channel]
	if seekFrom > 0 {
		// it must be -1, because start position here is excluded from array
		// https://github.com/s7techlab/hlf-sdk-go/blob/master/proto/seek.go#L15
		seekFrom--
	}

	peerChannel := &BlockPeerChannel{}
	peerChannel.Observer = NewBlockChannel(
		channel,
		bp.blockDeliverer,
		ChannelSeekFrom(seekFrom),
		WithChannelBlockLogger(bp.logger),
		WithChannelStopRecreateStream(bp.stopRecreateStream))

	_, peerChannel.err = peerChannel.Observer.Observe(ctx)
	if peerChannel.err != nil {
		bp.logger.Warn(`init channel Observer`, zap.Error(peerChannel.err))
	}

	// channel merger
	go func() {
		for b := range peerChannel.Observer.blocks {
			bp.blocks <- b
		}

		// after all reads peerParsedChannel.Observer.blocks close channels
		close(bp.blocks)
		for _, blocks := range bp.blocksByChannels {
			close(blocks)
		}
	}()

	return peerChannel
}
