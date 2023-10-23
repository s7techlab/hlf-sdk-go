package discovery

import (
	"context"
	"fmt"

	discoveryclient "github.com/hyperledger/fabric/discovery/client"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	grpcclient "github.com/s7techlab/hlf-sdk-go/client/grpc"
)

// implementation of api.DiscoveryProvider interface
var _ api.DiscoveryProvider = (*GossipDiscoveryProvider)(nil)

type GossipDiscoveryProvider struct {
	sd        *gossipServiceDiscovery
	tlsMapper connectionMapper
}

// return tls config for peers found via gossip
type connectionMapper interface {
	MapConnection(address string) *api.Endpoint
}

func NewGossipDiscoveryProvider(
	ctx context.Context,
	connCfg config.ConnectionConfig,
	log *zap.Logger,
	identitySigner discoveryclient.Signer,
	clientIdentity []byte,
	tlsMapper connectionMapper,
) (*GossipDiscoveryProvider, error) {
	discoveryClient, err := newFabricDiscoveryClient(ctx, connCfg, log, identitySigner)
	if err != nil {
		return nil, err
	}

	// TODO: probably we need to make a test call(ping) here to make sure user provided valid identity
	sd := newGossipServiceDiscovery(discoveryClient, clientIdentity)

	return &GossipDiscoveryProvider{sd: sd, tlsMapper: tlsMapper}, nil
}

// newFabricDiscoveryClient - initializes grpc fabric discovery client
// necessary for GossipDiscoveryProvider
func newFabricDiscoveryClient(
	ctx context.Context,
	c config.ConnectionConfig,
	log *zap.Logger,
	identitySigner discoveryclient.Signer,
) (*discoveryclient.Client, error) {
	dialOpts, err := grpcclient.OptionsFromConfig(c, log)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.DialContext(ctx, c.Host, dialOpts.Dial...)
	if err != nil {
		return nil, fmt.Errorf(`grpc dial to host=%s: %w`, c.Host, err)
	}

	discoveryClient := discoveryclient.NewClient(
		func() (*grpc.ClientConn, error) {
			return conn, nil
		},
		identitySigner,
		10,
	)

	return discoveryClient, nil
}

func (d *GossipDiscoveryProvider) Chaincode(ctx context.Context, channelName string, ccName string) (api.ChaincodeDiscoverer, error) {
	ccDTO, err := d.sd.DiscoverChaincode(ctx, ccName, channelName)
	if err != nil {
		return nil, err
	}

	return newChaincodeDiscovererTLSDecorator(ccDTO, d.tlsMapper), nil
}

func (d *GossipDiscoveryProvider) Channel(ctx context.Context, channelName string) (api.ChannelDiscoverer, error) {
	chanDTO, err := d.sd.DiscoverChannel(ctx, channelName)
	if err != nil {
		return nil, err
	}

	return newChannelDiscovererTLSDecorator(chanDTO, d.tlsMapper), nil
}

func (d *GossipDiscoveryProvider) LocalPeers(ctx context.Context) (api.LocalPeersDiscoverer, error) {
	localPeers, err := d.sd.LocalDiscovery(ctx)
	if err != nil {
		return nil, err
	}

	return newLocalPeersDiscovererTLSDecorator(localPeers, d.tlsMapper), nil
}
