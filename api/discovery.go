package api

import (
	"context"
)

type DiscoveryProvider interface {
	Chaincode(ctx context.Context, channelName string, ccName string) (ChaincodeDiscoverer, error)
	Channel(ctx context.Context, channelName string) (ChannelDiscoverer, error)
}

// ChaincodeDiscoverer - looking for info about network, channel, chaincode in local configs or gossip
type ChaincodeDiscoverer interface {
	Endorsers() []*HostEndpoint
	ChaincodeName() string
	ChaincodeVersion() string

	ChannelDiscoverer
}

// ChannelDiscoverer - info about orderers in channel
type ChannelDiscoverer interface {
	Orderers() []*HostEndpoint
	ChannelName() string
}

type HostEndpoint struct {
	MspID         string
	HostAddresses []string
}
