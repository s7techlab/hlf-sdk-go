package observer

import (
	"context"
	"fmt"
	"sync"

	"github.com/hyperledger/fabric-protos-go/common"
	"go.uber.org/zap"

	hlfproto "github.com/s7techlab/hlf-sdk-go/proto"
)

type (
	ParsedBlockChannel struct {
		blockChannel *BlockChannel

		transformers []BlockTransformer
		configBlock  *common.Block

		blocks        chan *ParsedBlock
		isWork        bool
		cancelObserve context.CancelFunc
		mutex         sync.Mutex
	}

	ParsedBlockChannelOpt func(*ParsedBlockChannel)
)

func WithParsedChannelBlockTransformers(transformers []BlockTransformer) ParsedBlockChannelOpt {
	return func(pbc *ParsedBlockChannel) {
		pbc.transformers = transformers
	}
}

func WithParsedChannelConfigBlock(configBlock *common.Block) ParsedBlockChannelOpt {
	return func(pbc *ParsedBlockChannel) {
		pbc.configBlock = configBlock
	}
}

func NewParsedBlockChannel(blockChannel *BlockChannel, opts ...ParsedBlockChannelOpt) *ParsedBlockChannel {
	parsedBlockChannel := &ParsedBlockChannel{
		blockChannel: blockChannel,
	}

	for _, opt := range opts {
		opt(parsedBlockChannel)
	}

	return parsedBlockChannel
}

func (p *ParsedBlockChannel) Observe(ctx context.Context) (<-chan *ParsedBlock, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.isWork {
		return p.blocks, nil
	}

	// ctxObserve using for nested control process without stopped primary context
	ctxObserve, cancel := context.WithCancel(ctx)
	p.cancelObserve = cancel

	incomingBlocks, err := p.blockChannel.Observe(ctxObserve)
	if err != nil {
		return nil, fmt.Errorf("observe common blocks: %w", err)
	}

	go func() {
		p.isWork = true

		for {
			select {
			case incomingBlock, hasMore := <-incomingBlocks:
				if !hasMore {
					continue
				}

				if incomingBlock == nil {
					continue
				}

				block := &ParsedBlock{
					Channel: p.blockChannel.channel,
				}
				block.Block, block.Error = hlfproto.ParseBlock(incomingBlock.Block, hlfproto.WithConfigBlock(p.configBlock))

				for pos, transformer := range p.transformers {
					if err = transformer.Transform(block); err != nil {
						p.blockChannel.logger.Warn(`transformer`, zap.Int(`pos`, pos), zap.Error(err))
					}
				}

				p.blocks <- block

			case <-ctxObserve.Done():
				p.Stop()
				return
			}
		}
	}()

	return p.blocks, nil
}

func (p *ParsedBlockChannel) Stop() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// If primary context is done then cancel ctxObserver
	if p.cancelObserve != nil {
		p.cancelObserve()
	}

	p.isWork = false
}
