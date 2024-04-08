package observer

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type PeerChannelsFetcherMock struct {
	channels map[string]uint64
}

func NewPeerChannelsFetcherMock(channels map[string]uint64) *PeerChannelsFetcherMock {
	return &PeerChannelsFetcherMock{channels: channels}
}

func (p *PeerChannelsFetcherMock) Uri() string {
	return "mock"
}

func (p *PeerChannelsFetcherMock) GetChannels(context.Context) (*peer.ChannelQueryResponse, error) {
	var channels []*peer.ChannelInfo
	for channelName := range p.channels {
		channels = append(channels, &peer.ChannelInfo{ChannelId: channelName})
	}

	return &peer.ChannelQueryResponse{Channels: channels}, nil
}

func (p *PeerChannelsFetcherMock) GetChainInfo(_ context.Context, channel string) (*common.BlockchainInfo, error) {
	chHeight, exists := p.channels[channel]
	if !exists {
		return nil, fmt.Errorf("channel '%s' does not exist", channel)
	}

	return &common.BlockchainInfo{Height: chHeight}, nil
}
