package observer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-protos-go/common"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/s7techlab/hlf-sdk-go/api"
)

const DefaultChannelPeerObservePeriod = 30 * time.Second

type (
	ChannelInfo struct {
		Channel   string
		Height    uint64
		UpdatedAt *timestamppb.Timestamp
	}

	// ChannelPeer observes for peer channels
	ChannelPeer struct {
		channelFetcher PeerChannelsFetcher

		channelsMatcher *ChannelsMatcher

		channels      map[string]*ChannelInfo
		observePeriod time.Duration

		lastError error
		mu        sync.Mutex
		logger    *zap.Logger

		isWork        bool
		cancelObserve context.CancelFunc
	}

	PeerChannelsFetcher interface {
		Uri() string
		api.ChannelListGetter
		api.ChainInfoGetter
	}

	PeerChannels interface {
		Uri() string
		Channels() map[string]*ChannelInfo
	}

	ChannelPeerOpts struct {
		channels      []ChannelToMatch
		observePeriod time.Duration
		logger        *zap.Logger
	}

	ChannelPeerOpt func(*ChannelPeerOpts)
)

var DefaultChannelPeerOpts = &ChannelPeerOpts{
	channels:      MatchAllChannels,
	observePeriod: DefaultChannelPeerObservePeriod,
	logger:        zap.NewNop(),
}

func WithChannels(channels []ChannelToMatch) ChannelPeerOpt {
	return func(opts *ChannelPeerOpts) {
		opts.channels = channels
	}
}

func WithChannelPeerLogger(logger *zap.Logger) ChannelPeerOpt {
	return func(opts *ChannelPeerOpts) {
		opts.logger = logger
	}
}

func NewChannelPeer(peerChannelsFetcher PeerChannelsFetcher, opts ...ChannelPeerOpt) (*ChannelPeer, error) {
	channelPeerOpts := DefaultChannelPeerOpts
	for _, opt := range opts {
		opt(channelPeerOpts)
	}

	channelsMatcher, err := NewChannelsMatcher(channelPeerOpts.channels)
	if err != nil {
		return nil, fmt.Errorf(`channels matcher: %w`, err)
	}

	channelPeer := &ChannelPeer{
		channelFetcher:  peerChannelsFetcher,
		channelsMatcher: channelsMatcher,
		channels:        make(map[string]*ChannelInfo),
		observePeriod:   channelPeerOpts.observePeriod,
		logger:          channelPeerOpts.logger,
	}

	return channelPeer, nil
}

func (cp *ChannelPeer) Stop() {
	cp.cancelObserve()
	cp.isWork = false
}

func (cp *ChannelPeer) Observe(ctx context.Context) {
	if cp.isWork {
		return
	}

	// ctxObserve using for nested control process without stopped primary context
	ctxObserve, cancel := context.WithCancel(context.Background())
	cp.cancelObserve = cancel

	go func() {
		cp.isWork = true
		cp.updateChannels(ctxObserve)

		ticker := time.NewTicker(cp.observePeriod)
		for {
			select {
			case <-ctx.Done():
				// If primary context is done then cancel ctxObserver
				cp.cancelObserve()
				return

			case <-ctxObserve.Done():
				return

			case <-ticker.C:
				cp.updateChannels(ctxObserve)
			}
		}
	}()
}

func (cp *ChannelPeer) Host() string {
	return cp.channelFetcher.Uri()
}

func (cp *ChannelPeer) Channels() map[string]*ChannelInfo {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	var copyChannelInfo = make(map[string]*ChannelInfo, len(cp.channels))
	for key, value := range cp.channels {
		copyChannelInfo[key] = value
	}

	return copyChannelInfo
}

func (cp *ChannelPeer) updateChannels(ctx context.Context) {
	cp.logger.Debug(`fetching channels`)
	channelsInfo, err := cp.channelFetcher.GetChannels(ctx)
	if err != nil {
		cp.logger.Warn(`error while fetching channels`, zap.Error(err))
		cp.lastError = err
		return
	}

	channels := ChannelsInfoToStrings(channelsInfo.Channels)
	cp.logger.Debug(`channels fetched`, zap.Strings(`channels`, channels))

	channelsMatched, err := cp.channelsMatcher.Match(channels)
	if err != nil {
		cp.logger.Warn(`channel matching error`, zap.Error(err))
		cp.lastError = err
		return
	}
	cp.logger.Debug(`channels matched`, zap.Reflect(`channels`, channelsMatched))

	channelHeights := make(map[string]uint64)

	for _, channel := range channelsMatched {
		var channelInfo *common.BlockchainInfo
		channelInfo, err = cp.channelFetcher.GetChainInfo(ctx, channel.Name)
		if err != nil {
			cp.lastError = err
			continue
		}
		channelHeights[channel.Name] = channelInfo.Height
	}

	cp.mu.Lock()
	defer cp.mu.Unlock()

	for channel, height := range channelHeights {
		var updatedAt *timestamp.Timestamp
		updatedAt, err = ptypes.TimestampProto(time.Now())
		if err != nil {
			cp.lastError = err
		}

		cp.channels[channel] = &ChannelInfo{
			Channel:   channel,
			Height:    height,
			UpdatedAt: updatedAt,
		}
	}
}
