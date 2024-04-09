package observer

import (
	"context"
	"strconv"
	"sync"
)

type Stream[T any] interface {
	Subscribe() (ch chan *Block[T], closer func())
}

type BlocksStream[T any] struct {
	connections map[string]chan *Block[T]
	mu          *sync.RWMutex

	isWork        bool
	cancelObserve context.CancelFunc
}

func NewBlocksStream[T any]() *BlocksStream[T] {
	return &BlocksStream[T]{
		connections: make(map[string]chan *Block[T]),
		mu:          &sync.RWMutex{},
	}
}

func (b *BlocksStream[T]) Observe(ctx context.Context, blocks <-chan *Block[T]) {
	if b.isWork {
		return
	}

	// ctxObserve using for nested control process without stopped primary context
	ctxObserve, cancel := context.WithCancel(ctx)
	b.cancelObserve = cancel

	go func() {
		defer func() {
			for connName := range b.connections {
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
				for _, connection := range b.connections {
					connection <- block
				}
				b.mu.RUnlock()
			}
		}
	}()
}

func (b *BlocksStream[T]) Subscribe() (chan *Block[T], func()) {
	b.mu.Lock()
	newConnection := make(chan *Block[T])
	name := "channel-" + strconv.Itoa(len(b.connections))
	b.connections[name] = newConnection
	b.mu.Unlock()

	closer := func() { b.closeChannel(name) }

	return newConnection, closer
}

func (b *BlocksStream[T]) closeChannel(name string) {
	b.mu.Lock()
	close(b.connections[name])
	delete(b.connections, name)
	b.mu.Unlock()
}

func (b *BlocksStream[T]) Stop() {
	if b.cancelObserve != nil {
		b.cancelObserve()
	}
	b.isWork = false
}
