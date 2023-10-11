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

		parsedChannelObservers map[string]*parsedBlockPeerChannel

		isWork        bool
		cancelObserve context.CancelFunc
	}

	parsedBlockPeerChannel struct {
		observer *ParsedBlockChannel
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
		parsedChannelObservers: make(map[string]*parsedBlockPeerChannel),
		blocks:                 make(chan *ParsedBlock),
		blocksByChannels:       make(map[string]chan *ParsedBlock),
	}

	for _, opt := range opts {
		opt(parsedBlockPeer)
	}

	return parsedBlockPeer
}

func (pbp *ParsedBlockPeer) ChannelObservers() map[string]*parsedBlockPeerChannel {
	pbp.mu.RLock()
	defer pbp.mu.RUnlock()

	var copyChannelObservers = make(map[string]*parsedBlockPeerChannel, len(pbp.parsedChannelObservers))
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

	pbp.blockPeer.Stop()

	for _, c := range pbp.parsedChannelObservers {
		if err := c.observer.Stop(); err != nil {
			zap.Error(err)
		}
	}

	pbp.parsedChannelObservers = make(map[string]*parsedBlockPeerChannel)

	if pbp.cancelObserve != nil {
		pbp.cancelObserve()
	}

	close(pbp.blocks)
	for _, blocksByChannel := range pbp.blocksByChannels {
		close(blocksByChannel)
	}

	pbp.isWork = false
}

func (pbp *ParsedBlockPeer) initParsedChannels(ctx context.Context) {
	pbp.mu.RLock()
	defer pbp.mu.RUnlock()

	for channel := range pbp.blockPeer.peerChannels.Channels() {
		if _, ok := pbp.parsedChannelObservers[channel]; !ok {
			pbp.blockPeer.logger.Info(`add parsed channel observer`, zap.String(`channel`, channel))

			pbp.parsedChannelObservers[channel] = pbp.peerParsedChannel(ctx, channel)
		}
	}
}

func (pbp *ParsedBlockPeer) peerParsedChannel(ctx context.Context, channel string) *parsedBlockPeerChannel {
	seekFrom := pbp.blockPeer.seekFrom[channel]
	if seekFrom > 0 {
		seekFrom--
	}

	commonBlockChannel := NewBlockChannel(channel, pbp.blockPeer.blockDeliverer, ChannelSeekFrom(seekFrom),
		WithChannelBlockLogger(pbp.blockPeer.logger), WithChannelStopRecreateStream(pbp.blockPeer.stopRecreateStream))

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

	// channel merger
	go func() {
		for b := range peerParsedChannel.observer.blocks {
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
