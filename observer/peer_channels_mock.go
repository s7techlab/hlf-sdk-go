package observer

import (
	"sync"
)

type PeerChannelsMock struct {
	mu           sync.Mutex
	channelsInfo map[string]*ChannelInfo
}

func NewPeerChannelsMock(channelsInfo ...*ChannelInfo) *PeerChannelsMock {
	channels := make(map[string]*ChannelInfo, len(channelsInfo))
	for _, channelInfo := range channelsInfo {
		channels[channelInfo.Channel] = channelInfo
	}

	return &PeerChannelsMock{channelsInfo: channels}
}

func (p *PeerChannelsMock) URI() string {
	return "mock"
}

func (p *PeerChannelsMock) Channels() map[string]*ChannelInfo {
	p.mu.Lock()
	defer p.mu.Unlock()

	var copyChannelInfo = make(map[string]*ChannelInfo, len(p.channelsInfo))
	for key, value := range p.channelsInfo {
		copyChannelInfo[key] = value
	}

	return copyChannelInfo
}

func (p *PeerChannelsMock) UpdateChannelInfo(channelInfo *ChannelInfo) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.channelsInfo[channelInfo.Channel] = channelInfo
}
