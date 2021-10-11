package discovery

import (
	"context"

	discClient "github.com/hyperledger/fabric/discovery/client"
	"github.com/s7techlab/hlf-sdk-go/api"
)

type GossipDiscoveryProvider struct {
	sd *gossipServiceDiscovery
}

func NewGossipDiscoveryProvider(client *discClient.Client, clientIdentity []byte) *GossipDiscoveryProvider {
	sd := newGossipServiceDiscovery(client, clientIdentity)

	return &GossipDiscoveryProvider{sd}
}

func (d *GossipDiscoveryProvider) Chaincode(ctx context.Context, channelName string, ccName string) (api.IDiscoveryChaincode, error) {
	return d.sd.Discover(ctx, ccName, channelName)
}
