package subs

import (
	"github.com/hyperledger/fabric/protos/common"
)

type (
	// BlockHandler  when block == nil is eq EOF and signal for terminate all sub channels
	BlockHandler func(block *common.Block) bool

	ErrorCloser interface {
		Err() <-chan error
		Errors() chan error
		Close() error
	}
)

func NewBlockSubscription() *BlockSubscription {
	return &BlockSubscription{
		blocks: make(chan *common.Block, 0),
	}
}

type BlockSubscription struct {
	blocks chan *common.Block
	ErrorCloser
}

func (b *BlockSubscription) Blocks() <-chan *common.Block {
	return b.blocks
}

func (b *BlockSubscription) Handler(block *common.Block) bool {
	if block == nil {
		close(b.blocks)
	} else {
		b.blocks <- block
	}

	return false
}

func (b *BlockSubscription) Serve(base ErrorCloser) *BlockSubscription {
	b.ErrorCloser = base
	return b
}
