package observer

import (
	"context"
	"strconv"
	"sync"
)

type StreamParsed interface {
	Subscribe() (ch chan *ParsedBlock, closer func())
}

type ParsedBlocksStream struct {
	connectionsParsed map[string]chan *ParsedBlock
	mu                *sync.RWMutex

	isWork        bool
	cancelObserve context.CancelFunc
}

func NewParsedBlocksStream() *ParsedBlocksStream {
	return &ParsedBlocksStream{
		connectionsParsed: make(map[string]chan *ParsedBlock),
		mu:                &sync.RWMutex{},
	}
}

func (b *ParsedBlocksStream) Observe(ctx context.Context, blocks <-chan *ParsedBlock) {
	if b.isWork {
		return
	}

	// ctxObserve using for nested control process without stopped primary context
	ctxObserve, cancel := context.WithCancel(ctx)
	b.cancelObserve = cancel

	go func() {
		defer func() {
			for connName := range b.connectionsParsed {
				b.closeChannel(connName)
			}
		}()

		b.isWork = true

		for {
			select {
			case <-ctxObserve.Done():
				// If primary context is done then cancel ctxObserver
				b.cancelObserve()
				return

			case block, ok := <-blocks:
				if !ok {
					return
				}

				b.mu.RLock()
				for _, connection := range b.connectionsParsed {
					connection <- block
				}
				b.mu.RUnlock()
			}
		}
	}()
}

func (b *ParsedBlocksStream) Subscribe() (chan *ParsedBlock, func()) {
	b.mu.Lock()
	newConnection := make(chan *ParsedBlock)
	name := "channel-" + strconv.Itoa(len(b.connectionsParsed))
	b.connectionsParsed[name] = newConnection
	b.mu.Unlock()

	closer := func() { b.closeChannel(name) }

	return newConnection, closer
}

func (b *ParsedBlocksStream) SubscribeParsed() (chan *ParsedBlock, func()) {
	b.mu.Lock()
	newConnection := make(chan *ParsedBlock)
	name := "channel-" + strconv.Itoa(len(b.connectionsParsed))
	b.connectionsParsed[name] = newConnection
	b.mu.Unlock()

	closer := func() { b.closeChannel(name) }

	return newConnection, closer
}

func (b *ParsedBlocksStream) closeChannel(name string) {
	b.mu.Lock()
	close(b.connectionsParsed[name])
	delete(b.connectionsParsed, name)
	b.mu.Unlock()
}

func (b *ParsedBlocksStream) Stop() {
	if b.cancelObserve != nil {
		b.cancelObserve()
	}
	b.isWork = false
}
