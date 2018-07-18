package api

import (
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/s7techlab/hlf-sdk-go/api/config"
)

const (
	CCTypeGoLang = `golang`
)

type DiscoveryProviderOpts map[string]interface{}

type DiscoveryProvider interface {
	Initialize(opts config.DiscoveryConfigOpts) (DiscoveryProvider, error)
	Channels() ([]DiscoveryChannel, error)
	Chaincode(channelName string, ccName string) (*DiscoveryChaincode, error)
	Chaincodes(channelName string) ([]DiscoveryChaincode, error)
	Endorsers(channelName string, ccName string) ([]config.PeerConfig, error)
}

type DiscoveryChannel struct {
	Name        string               `json:"channel_name" yaml:"name"`
	Description string               `json:"channel_description" yaml:"description"`
	Chaincodes  []DiscoveryChaincode `json:"chaincodes" yaml:"description"`
	Peers       []config.PeerConfig  `json:"peers" yaml:"peers"`
}

type DiscoveryChaincode struct {
	Name        string `json:"chaincode_name" yaml:""`
	Type        string `json:"type"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

func (c DiscoveryChaincode) GetFabricType() peer.ChaincodeSpec_Type {
	switch c.Type {
	case CCTypeGoLang:
		return peer.ChaincodeSpec_GOLANG
	}
	return peer.ChaincodeSpec_UNDEFINED
}
