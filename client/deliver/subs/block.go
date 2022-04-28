package subs

import (
	"github.com/hyperledger/fabric-protos-go/common"
)

type (
	// BlockHandler  when block == nil is eq EOF and signal for terminate all sub channels
	BlockHandler     func(block *common.Block) bool
	ReadyForHandling func()

	ErrorCloser interface {
		Done() <-chan struct{}
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
		select {
		case b.blocks <- block:
		case <-b.ErrorCloser.Done():
			return true
		}
	}

	return false
}

func (b *BlockSubscription) Serve(base ErrorCloser, readyForHandling ReadyForHandling) *BlockSubscription {
	b.ErrorCloser = base
	readyForHandling()
	return b
}
