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

const DefaultPeerChannelsObservePeriod = 30 * time.Second

type (
	ChannelInfo struct {
		Channel   string
		Height    uint64
		UpdatedAt *timestamppb.Timestamp
	}

	// PeerChannels observes for peer channels
	PeerChannels struct {
		channels map[string]*ChannelInfo

		channelFetcher  PeerChannelsFetcher
		channelsMatcher *ChannelsMatcher
		observePeriod   time.Duration

		mu     sync.Mutex
		logger *zap.Logger

		lastError error

		isWork        bool
		cancelObserve context.CancelFunc
	}

	PeerChannelsFetcher interface {
		Uri() string
		api.ChannelListGetter
		api.ChainInfoGetter
	}

	PeerChannelsOpts struct {
		channels      []ChannelToMatch
		observePeriod time.Duration
		logger        *zap.Logger
	}

	PeerChannelsOpt func(*PeerChannelsOpts)
)

var DefaultPeerChannelsOpts = &PeerChannelsOpts{
	channels:      MatchAllChannels,
	observePeriod: DefaultPeerChannelsObservePeriod,
	logger:        zap.NewNop(),
}

func WithChannels(channels []ChannelToMatch) PeerChannelsOpt {
	return func(opts *PeerChannelsOpts) {
		opts.channels = channels
	}
}

func WithPeerChannelsLogger(logger *zap.Logger) PeerChannelsOpt {
	return func(opts *PeerChannelsOpts) {
		opts.logger = logger
	}
}

func NewPeerChannels(peerChannelsFetcher PeerChannelsFetcher, opts ...PeerChannelsOpt) (*PeerChannels, error) {
	peerChannelsOpts := DefaultPeerChannelsOpts
	for _, opt := range opts {
		opt(peerChannelsOpts)
	}

	channelsMatcher, err := NewChannelsMatcher(peerChannelsOpts.channels)
	if err != nil {
		return nil, fmt.Errorf(`channels matcher: %w`, err)
	}

	peerChannels := &PeerChannels{
		channelFetcher:  peerChannelsFetcher,
		channelsMatcher: channelsMatcher,
		channels:        make(map[string]*ChannelInfo),
		observePeriod:   peerChannelsOpts.observePeriod,
		logger:          peerChannelsOpts.logger,
	}

	return peerChannels, nil
}

func (pc *PeerChannels) Stop() {
	pc.cancelObserve()
	pc.isWork = false
}

func (pc *PeerChannels) Observe(ctx context.Context) {
	if pc.isWork {
		return
	}

	// ctxObserve using for nested control process without stopped primary context
	ctxObserve, cancel := context.WithCancel(context.Background())
	pc.cancelObserve = cancel

	go func() {
		pc.isWork = true
		pc.updateChannels(ctxObserve)

		ticker := time.NewTicker(pc.observePeriod)
		for {
			select {
			case <-ctx.Done():
				// If primary context is done then cancel ctxObserver
				pc.cancelObserve()
				return

			case <-ctxObserve.Done():
				return

			case <-ticker.C:
				pc.updateChannels(ctxObserve)
			}
		}
	}()
}

func (pc *PeerChannels) Uri() string {
	return pc.channelFetcher.Uri()
}

func (pc *PeerChannels) Channels() map[string]*ChannelInfo {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	var copyChannelInfo = make(map[string]*ChannelInfo, len(pc.channels))
	for key, value := range pc.channels {
		copyChannelInfo[key] = value
	}

	return copyChannelInfo
}

func (pc *PeerChannels) updateChannels(ctx context.Context) {
	pc.logger.Debug(`fetching channels`)
	channelsInfo, err := pc.channelFetcher.GetChannels(ctx)
	if err != nil {
		pc.logger.Warn(`error while fetching channels`, zap.Error(err))
		pc.lastError = err
		return
	}

	channels := ChannelsInfoToStrings(channelsInfo.Channels)
	pc.logger.Debug(`channels fetched`, zap.Strings(`channels`, channels))

	channelsMatched, err := pc.channelsMatcher.Match(channels)
	if err != nil {
		pc.logger.Warn(`channel matching error`, zap.Error(err))
		pc.lastError = err
		return
	}
	pc.logger.Debug(`channels matched`, zap.Reflect(`channels`, channelsMatched))

	channelHeights := make(map[string]uint64)

	for _, channel := range channelsMatched {
		var channelInfo *common.BlockchainInfo
		channelInfo, err = pc.channelFetcher.GetChainInfo(ctx, channel.Name)
		if err != nil {
			pc.lastError = err
			continue
		}
		channelHeights[channel.Name] = channelInfo.Height
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	for channel, height := range channelHeights {
		var updatedAt *timestamp.Timestamp
		updatedAt, err = ptypes.TimestampProto(time.Now())
		if err != nil {
			pc.lastError = err
		}

		pc.channels[channel] = &ChannelInfo{
			Channel:   channel,
			Height:    height,
			UpdatedAt: updatedAt,
		}
	}
}
