package discovery

import (
	"context"
	"fmt"

	discClient "github.com/hyperledger/fabric/discovery/client"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/util"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// implementation of api.DiscoveryProvider interface
var _ api.DiscoveryProvider = (*GossipDiscoveryProvider)(nil)

type GossipDiscoveryProvider struct {
	sd *gossipServiceDiscovery
}

func NewGossipDiscoveryProvider(
	ctx context.Context,
	connCfg config.ConnectionConfig,
	log *zap.Logger,
	identitySigner discClient.Signer,
	clientIdentity []byte,
) (*GossipDiscoveryProvider, error) {
	discClient, err := newFabricDiscoveryClient(ctx, connCfg, log, identitySigner)
	if err != nil {
		return nil, err
	}

	// TODO probably we need to make a test call(ping) here to make sure user provided valid identity
	sd := newGossipServiceDiscovery(discClient, clientIdentity)

	return &GossipDiscoveryProvider{sd}, nil
}

// newFabricDiscoveryClient - initializes grpc fabric discovery client
// necessary for GossipDiscoveryProvider
func newFabricDiscoveryClient(
	ctx context.Context,
	c config.ConnectionConfig,
	log *zap.Logger,
	identitySigner discClient.Signer,
) (*discClient.Client, error) {
	opts, err := util.NewGRPCOptionsFromConfig(c, log)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.DialContext(ctx, c.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf(`grpc dial to host=%s: %w`, c.Host, err)
	}

	discClient := discClient.NewClient(
		func() (*grpc.ClientConn, error) {
			return conn, nil
		},
		identitySigner,
		10,
	)

	return discClient, nil
}

func (d *GossipDiscoveryProvider) Chaincode(ctx context.Context, channelName string, ccName string) (api.ChaincodeDiscoverer, error) {
	return d.sd.DiscoverChaincode(ctx, ccName, channelName)
}

func (d *GossipDiscoveryProvider) Channel(ctx context.Context, channelName string) (api.ChannelDiscoverer, error) {
	return d.sd.DiscoverChannel(ctx, channelName)
}
