package observer

import (
	"context"
	"sync"
	"time"

	"github.com/hyperledger/fabric-protos-go/common"
	"go.uber.org/zap"
)

type (
	ParsedBlockPeer struct {
		mu sync.RWMutex

		blockPeer    *BlockPeer
		transformers []BlockTransformer
		configBlocks map[string]*common.Block

		blocks           chan *ParsedBlock
		blocksByChannels map[string]chan *ParsedBlock

		parsedChannelObservers map[string]*ParsedBlockPeerChannel

		isWork        bool
		cancelObserve context.CancelFunc
	}

	ParsedBlockPeerChannel struct {
		Observer *ParsedBlockChannel
		err      error
	}

	ParsedBlockPeerOpt func(*ParsedBlockPeer)
)

func WithBlockPeerTransformer(transformers ...BlockTransformer) ParsedBlockPeerOpt {
	return func(pbp *ParsedBlockPeer) {
		pbp.transformers = transformers
	}
}

// WithConfigBlocks just for correct parsing of BFT at hlfproto.ParseBlock
func WithConfigBlocks(configBlocks map[string]*common.Block) ParsedBlockPeerOpt {
	return func(pbp *ParsedBlockPeer) {
		pbp.configBlocks = configBlocks
	}
}

func NewParsedBlockPeer(blocksPeer *BlockPeer, opts ...ParsedBlockPeerOpt) *ParsedBlockPeer {
	parsedBlockPeer := &ParsedBlockPeer{
		blockPeer:              blocksPeer,
		parsedChannelObservers: make(map[string]*ParsedBlockPeerChannel),
		blocks:                 make(chan *ParsedBlock),
		blocksByChannels:       make(map[string]chan *ParsedBlock),
	}

	for _, opt := range opts {
		opt(parsedBlockPeer)
	}

	return parsedBlockPeer
}

func (pbp *ParsedBlockPeer) ChannelObservers() map[string]*ParsedBlockPeerChannel {
	pbp.mu.RLock()
	defer pbp.mu.RUnlock()

	var copyChannelObservers = make(map[string]*ParsedBlockPeerChannel, len(pbp.parsedChannelObservers))
	for key, value := range pbp.parsedChannelObservers {
		copyChannelObservers[key] = value
	}

	return copyChannelObservers
}

func (pbp *ParsedBlockPeer) Observe(ctx context.Context) <-chan *ParsedBlock {
	if pbp.isWork {
		return pbp.blocks
	}

	// ctxObserve using for nested control process without stopped primary context
	ctxObserve, cancel := context.WithCancel(ctx)
	pbp.cancelObserve = cancel

	pbp.initParsedChannels(ctxObserve)

	// init new channels if they are fetched
	go func() {
		pbp.isWork = true

		time.Sleep(time.Second)

		ticker := time.NewTicker(pbp.blockPeer.observePeriod)
		for {
			select {
			case <-ctxObserve.Done():
				pbp.Stop()
				return

			case <-ticker.C:
				pbp.initParsedChannels(ctxObserve)
			}
		}
	}()

	return pbp.blocks
}

func (pbp *ParsedBlockPeer) Stop() {
	pbp.mu.Lock()
	defer pbp.mu.Unlock()

	// pbp.blocks and pbp.blocksByChannels mustn't be closed here, because they are closed elsewhere

	pbp.blockPeer.Stop()

	for _, c := range pbp.parsedChannelObservers {
		if err := c.Observer.Stop(); err != nil {
			zap.Error(err)
		}
	}

	pbp.parsedChannelObservers = make(map[string]*ParsedBlockPeerChannel)

	if pbp.cancelObserve != nil {
		pbp.cancelObserve()
	}

	pbp.isWork = false
}

func (pbp *ParsedBlockPeer) initParsedChannels(ctx context.Context) {
	for channel := range pbp.blockPeer.peerChannels.Channels() {
		pbp.mu.RLock()
		_, ok := pbp.parsedChannelObservers[channel]
		pbp.mu.RUnlock()

		if !ok {
			pbp.blockPeer.logger.Info(`add parsed channel observer`, zap.String(`channel`, channel))

			parsedBlockPeerChannel := pbp.peerParsedChannel(ctx, channel)

			pbp.mu.Lock()
			pbp.parsedChannelObservers[channel] = parsedBlockPeerChannel
			pbp.mu.Unlock()
		}
	}
}

func (pbp *ParsedBlockPeer) peerParsedChannel(ctx context.Context, channel string) *ParsedBlockPeerChannel {
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

	// channel merger
	go func() {
		for b := range peerParsedChannel.Observer.parsedBlocks {
			pbp.blocks <- b
		}

		// after all reads peerParsedChannel.observer.blocks close channels
		close(pbp.blocks)
		for _, blocks := range pbp.blocksByChannels {
			close(blocks)
		}
	}()

	return peerParsedChannel
}
