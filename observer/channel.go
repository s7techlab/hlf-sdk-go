package observer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hyperledger/fabric/msp"
	"go.uber.org/zap"
)

var ErrChannelObserverAlreadyStarted = errors.New(`channel observer already started`)

type (
	SeekFromFetcher func(ctx context.Context, channel string) (uint64, error)

	Opts struct {
		identity msp.SigningIdentity
		logger   *zap.Logger
	}

	Channel struct {
		// current name of channel
		channel string

		seekFromFetcher SeekFromFetcher

		identity msp.SigningIdentity

		// graceful shutdown
		closer func() error

		// value received from seekFromFetcher
		lastSeekFrom uint64

		// status current
		status ChannelObserverStatus
		// in case of error how many times we tried co reconnect
		connectAttempt   uint64
		connectAttemptAt time.Time
		// when we subscribed to channel
		connectedAt time.Time

		// number of last fetched observer
		// lastFetchedBlock uint64

		// last errors we got
		lastError error

		logger *zap.Logger

		mu sync.Mutex
	}

	ChannelObserverStatus int
)

const (
	ChannelObserverCreated ChannelObserverStatus = iota
	ChannelObserverConnecting
	ChannelObserverConnected
	ChannelObserverStopped
	ChannelObserverErrored

	DefaultConnectRetryDelay = 5 * time.Second
)

var DefaultOpts = &Opts{
	identity: nil,          // use default identity in blocksDeliverer
	logger:   zap.NewNop(), // silent logger
}

func (s ChannelObserverStatus) String() string {
	return [...]string{`Created`, `Connecting`, `Connected`, `Stopped`, `Errored`}[s]
}

func ChannelSeekFrom(seekFrom uint64) SeekFromFetcher {
	return func(ctx context.Context, channel string) (uint64, error) {
		return seekFrom, nil
	}
}

func ChannelSeekOldest() SeekFromFetcher {
	return ChannelSeekFrom(0)
}

func (c *Channel) setStatus(status ChannelObserverStatus) {
	c.status = status
}

func (c *Channel) allowToObserve() error {
	if c.status != ChannelObserverCreated && c.status != ChannelObserverStopped {
		return ErrChannelObserverAlreadyStarted
	}
	return nil
}

func (c *Channel) preCreateStream() {
	c.closer = nil
	c.connectAttempt++
	c.connectAttemptAt = time.Now()
	c.connectedAt = time.Unix(0, 0) // zero
	c.setStatus(ChannelObserverConnecting)
}

func (c *Channel) afterCreateStream(closer func() error) {
	c.closer = closer
	c.connectedAt = time.Now()
	c.setStatus(ChannelObserverConnected)
}

func (c *Channel) setError(err error) {
	c.lastError = err
	c.setStatus(ChannelObserverErrored)
}

func (c *Channel) processSeekFrom(ctx context.Context) (uint64, error) {
	seekFrom, err := c.seekFromFetcher(ctx, c.channel)
	if err != nil {
		c.setError(err)
		return 0, fmt.Errorf(`seek from: %w`, c.lastError)
	}

	c.lastSeekFrom = seekFrom
	return seekFrom, nil
}

func (c *Channel) stop() error {
	if c == nil {
		return nil
	}

	if c.closer != nil {
		// close incoming stream
		c.lastError = c.closer()
		c.closer = nil
	}

	c.setStatus(ChannelObserverStopped)
	return c.lastError
}
