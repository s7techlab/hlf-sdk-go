package observer

import (
	"sync"
)

type ChannelPeerMock struct {
	mu           sync.Mutex
	channelsInfo map[string]*ChannelInfo
}

func NewChannelPeerMock(channelsInfo ...*ChannelInfo) *ChannelPeerMock {
	channels := make(map[string]*ChannelInfo, len(channelsInfo))
	for _, channelInfo := range channelsInfo {
		channels[channelInfo.Channel] = channelInfo
	}

	return &ChannelPeerMock{
		channelsInfo: channels,
	}
}

func (m *ChannelPeerMock) Channels() map[string]*ChannelInfo {
	m.mu.Lock()
	defer m.mu.Unlock()

	var copyChannelInfo = make(map[string]*ChannelInfo, len(m.channelsInfo))
	for key, value := range m.channelsInfo {
		copyChannelInfo[key] = value
	}

	return copyChannelInfo
}

func (m *ChannelPeerMock) UpdateChannelInfo(channelInfo *ChannelInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.channelsInfo[channelInfo.Channel] = channelInfo
}
