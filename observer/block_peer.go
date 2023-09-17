package observer

import (
	"context"
	"sync"
	"time"

	"github.com/hyperledger/fabric-protos-go/common"
	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/api"
)

const DefaultBlockPeerObservePeriod = 10 * time.Second

type (
	BlockPeer struct {
		mu sync.RWMutex

		peerChannels       PeerChannels
		blockDeliverer     api.BlocksDeliverer
		channelObservers   map[string]*blockPeerChannel
		transformers       []BlockTransformer
		seekFrom           map[string]uint64
		configBlocks       map[string]*common.Block
		observePeriod      time.Duration
		stopRecreateStream bool
		logger             *zap.Logger

		blocks           chan *Block
		blocksByChannels map[string]chan *Block

		isWork        bool
		cancelObserve context.CancelFunc
	}

	blockPeerChannel struct {
		observer *BlockChannel
		err      error
	}

	BlockPeerOpts struct {
		transformers       []BlockTransformer
		seekFrom           map[string]uint64
		configBlocks       map[string]*common.Block
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

func WithBlockPeerTransformer(transformers ...BlockTransformer) BlockPeerOpt {
	return func(opts *BlockPeerOpts) {
		opts.transformers = transformers
	}
}

func WithSeekFrom(seekFrom map[string]uint64) BlockPeerOpt {
	return func(opts *BlockPeerOpts) {
		opts.seekFrom = seekFrom
	}
}

// WithConfigBlocks just for correct parsing of BFT at hlfproto.ParseBlock
func WithConfigBlocks(configBlocks map[string]*common.Block) BlockPeerOpt {
	return func(opts *BlockPeerOpts) {
		opts.configBlocks = configBlocks
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
		channelObservers:   make(map[string]*blockPeerChannel),
		blocks:             make(chan *Block),
		blocksByChannels:   make(map[string]chan *Block),
		transformers:       blockPeerOpts.transformers,
		seekFrom:           blockPeerOpts.seekFrom,
		configBlocks:       blockPeerOpts.configBlocks,
		observePeriod:      blockPeerOpts.observePeriod,
		stopRecreateStream: blockPeerOpts.stopRecreateStream,
		logger:             blockPeerOpts.logger,
	}

	return blockPeer
}

func (bp *BlockPeer) ChannelObservers() map[string]*blockPeerChannel {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	var copyChannelObservers = make(map[string]*blockPeerChannel, len(bp.channelObservers))
	for key, value := range bp.channelObservers {
		copyChannelObservers[key] = value
	}

	return copyChannelObservers
}

func (bp *BlockPeer) Observe(ctx context.Context) (<-chan *Block, error) {
	if bp.isWork {
		return bp.blocks, nil
	}

	// ctxObserve using for nested controll process without stopped primary context
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

	return bp.blocks, nil
}

func (bp *BlockPeer) Stop() {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	for _, c := range bp.channelObservers {
		if err := c.observer.Stop(); err != nil {
			zap.Error(err)
		}
	}

	bp.channelObservers = make(map[string]*blockPeerChannel)

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
			bp.logger.Info(`add channel observer`, zap.String(`channel`, channel))

			bp.channelObservers[channel] = bp.peerChannel(ctx, channel)
		}
	}
}

func (bp *BlockPeer) peerChannel(ctx context.Context, channel string) *blockPeerChannel {
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

	// channel merger
	go func() {
		for b := range peerChannel.observer.blocks {
			bp.blocks <- b
		}

		// after all reads peerChannel.observer.blocks close channels
		close(bp.blocks)
		for _, blocks := range bp.blocksByChannels {
			close(blocks)
		}
	}()

	return peerChannel
}
