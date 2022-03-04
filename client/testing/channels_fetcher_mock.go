package testing

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/peer"
)

type ChannelsFetcherMock struct {
	channels []string
}

func NewChannelsFetcherMock(channels []string) *ChannelsFetcherMock {
	return &ChannelsFetcherMock{
		channels: channels,
	}
}

func (c ChannelsFetcherMock) GetChannels(ctx context.Context) (*peer.ChannelQueryResponse, error) {
	var channels []*peer.ChannelInfo
	for i := range c.channels {
		channels = append(channels, &peer.ChannelInfo{ChannelId: c.channels[i]})
	}

	return &peer.ChannelQueryResponse{
		Channels: channels,
	}, nil
}
