package observer

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type ChannelPeerFetcherMock struct {
	channels map[string]uint64
}

func NewChannelPeerFetcherMock(channels map[string]uint64) *ChannelPeerFetcherMock {
	return &ChannelPeerFetcherMock{
		channels: channels,
	}
}

func (c *ChannelPeerFetcherMock) Uri() string {
	return "mock"
}

func (c *ChannelPeerFetcherMock) GetChannels(context.Context) (*peer.ChannelQueryResponse, error) {
	var channels []*peer.ChannelInfo
	for channelName := range c.channels {
		channels = append(channels, &peer.ChannelInfo{ChannelId: channelName})
	}

	return &peer.ChannelQueryResponse{
		Channels: channels,
	}, nil
}

func (c *ChannelPeerFetcherMock) GetChainInfo(_ context.Context, channel string) (*common.BlockchainInfo, error) {
	chHeight, exists := c.channels[channel]
	if !exists {
		return nil, fmt.Errorf("channel '%s' does not exist", channel)
	}

	return &common.BlockchainInfo{
		Height: chHeight,
	}, nil
}
