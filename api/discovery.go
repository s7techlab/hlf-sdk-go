package api

import (
	"context"

	"github.com/s7techlab/hlf-sdk-go/api/config"
)

type DiscoveryProvider interface {
	// rm Initialize(opts config.DiscoveryConfigOpts, pool PeerPool, core Core) (DiscoveryProvider, error)
	// rm Channels() ([]DiscoveryChannel, error)

	// ? Channel(channelName string) (*DiscoveryChannel, error)
	Chaincode(ctx context.Context, channelName string, ccName string) (IDiscoveryChaincode, error)
	// ? Chaincodes(channelName string) ([]DiscoveryChaincode, error)
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

type IDiscoveryChaincode interface {
	Endorsers() []*HostEndpoint
	Orderers() []*HostEndpoint
	ChaincodeName() string
	ChaincodeVersion() string
	ChannelName() string
}

type DiscoveryChaincode struct {
	Name        string `json:"chaincode_name" yaml:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Policy      string `json:"policy"`
}
