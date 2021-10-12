package api

import (
	"context"

	"github.com/s7techlab/hlf-sdk-go/api/config"
)

type DiscoveryProvider interface {
	Chaincode(ctx context.Context, channelName string, ccName string) (ChaincodeDiscoverer, error)
	Channel(ctx context.Context, channelName string) (ChannelDiscoverer, error)
}

// ChaincodeDiscoverer - looking for info about network in configs or gossip
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

type DiscoveryChannel struct {
	Name        string                    `json:"channel_name" yaml:"name"`
	Description string                    `json:"channel_description" yaml:"description"`
	Chaincodes  []DiscoveryChaincode      `json:"chaincodes" yaml:"description"`
	Orderers    []config.ConnectionConfig `json:"orderers" yaml:"orderers"`
}

type HostEndpoint struct {
	MspID         string
	HostAddresses []string
}

type DiscoveryChaincode struct {
	Name        string `json:"chaincode_name" yaml:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Policy      string `json:"policy"`
}
