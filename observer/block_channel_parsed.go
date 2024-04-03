package observer

import (
	"context"
	"sync"

	"github.com/hyperledger/fabric-protos-go/common"
	"go.uber.org/zap"

	hlfproto "github.com/s7techlab/hlf-sdk-go/block"
)

type (
	ParsedBlockChannel struct {
		BlockChannel *BlockChannel

		transformers []BlockTransformer
		configBlock  *common.Block

		parsedBlocks  chan *ParsedBlock
		isWork        bool
		cancelObserve context.CancelFunc
		mu            sync.Mutex
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
		BlockChannel: blockChannel,
	}

	for _, opt := range opts {
		opt(parsedBlockChannel)
	}

	return parsedBlockChannel
}

func (p *ParsedBlockChannel) Observe(ctx context.Context) (<-chan *ParsedBlock, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isWork {
		return p.parsedBlocks, nil
	}

	// ctxObserve using for nested control process without stopped primary context
	ctxObserve, cancel := context.WithCancel(ctx)
	p.cancelObserve = cancel

	if err := p.BlockChannel.allowToObserve(); err != nil {
		return nil, err
	}

	// Double check
	if err := p.BlockChannel.allowToObserve(); err != nil {
		return nil, err
	}

	p.parsedBlocks = make(chan *ParsedBlock)

	go func() {
		p.isWork = true

		p.BlockChannel.logger.Debug(`creating parsed block stream`)
		incomingBlocks, errCreateStream := p.BlockChannel.createParsedStreamWithRetry(ctxObserve, p.BlockChannel.createParsedStream)
		if errCreateStream != nil {
			return
		}

		p.BlockChannel.logger.Info(`parsed block stream created`)
		for {
			select {
			case incomingBlock, hasMore := <-incomingBlocks:

				var err error
				if !hasMore && !p.BlockChannel.stopRecreateStream {
					p.BlockChannel.logger.Debug(`parsed block stream interrupted, recreate`)
					incomingBlocks, err = p.BlockChannel.createParsedStreamWithRetry(ctx, p.BlockChannel.createParsedStream)
					if err != nil {
						return
					}

					p.BlockChannel.logger.Debug(`parsed block stream recreated`)
					continue
				}

				if incomingBlock == nil {
					continue
				}

				parsedBlock := &ParsedBlock{
					Channel:       p.BlockChannel.channel,
					BlockOriginal: incomingBlock,
					Block:         incomingBlock,
				}

				bftOrdererIdentities, err := hlfproto.ParseBTFOrderersIdentities(parsedBlock.BlockOriginal.GetMetadata().GetRawUnparsedMetadataSignatures(), p.configBlock)
				if err != nil {
					p.BlockChannel.logger.Error("parse bft orderers identities", zap.Error(err))
				}
				parsedBlock.Block.Metadata.OrdererSignatures = append(parsedBlock.Block.Metadata.OrdererSignatures, bftOrdererIdentities...)

				for pos, transformer := range p.transformers {
					if err = transformer.Transform(parsedBlock); err != nil {
						p.BlockChannel.logger.Warn(`transformer`, zap.Int(`pos`, pos), zap.Error(err))
					}
				}

				p.parsedBlocks <- parsedBlock

			case <-ctxObserve.Done():
				if err := p.Stop(); err != nil {
					p.BlockChannel.lastError = err
				}
				return
			}
		}
	}()

	return p.parsedBlocks, nil
}

func (p *ParsedBlockChannel) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// p.blocks mustn't be closed here, because it is closed elsewhere

	err := p.BlockChannel.Stop()

	// If primary context is done then cancel ctxObserver
	if p.cancelObserve != nil {
		p.cancelObserve()
	}

	p.isWork = false
	return err
}
