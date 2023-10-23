package client

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/hyperledger/fabric/msp"
	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/client/discovery"
	"github.com/s7techlab/hlf-sdk-go/client/grpc"
	"github.com/s7techlab/hlf-sdk-go/crypto"
)

// implementation of api.Core interface
var _ api.Client = (*Client)(nil)

var (
	ErrEmptyMSPConfig              = errors.New(`empty MSP config`)
	ErrDiscoveryConnectionRequired = errors.New(`discovery connection required`)
	ErrDiscoverySignerRequired     = errors.New(`discovery signer required`)
)

type Client struct {
	ctx    context.Context
	config *config.Config

	defaultSigner msp.SigningIdentity // default signer for requests

	peerPool api.PeerPool
	orderer  api.Orderer

	discoveryProvider api.DiscoveryProvider
	discoverySigner   msp.SigningIdentity // signer for discovery queries

	channels  map[string]api.Channel
	channelMx sync.Mutex

	crypto   crypto.CryptoSuite
	logger   *zap.Logger
	fabricV2 bool
}

func New(ctx context.Context, opts ...Opt) (*Client, error) {
	var err error

	client := &Client{
		ctx:      ctx,
		config:   &config.Config{},
		channels: make(map[string]api.Channel),
	}

	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, fmt.Errorf(`apply option: %w`, err)
		}
	}

	if err := applyDefaults(client); err != nil {
		return nil, err
	}

	// if peerPool is empty, set it from config
	if client.peerPool != nil {
		if err = client.initPeerPool(); err != nil {
			return nil, fmt.Errorf(`init peer pool: %w`, err)
		}
	}

	if client.discoveryProvider == nil && client.config != nil {
		mapper := discovery.NewEndpointsMapper(client.config.EndpointsMap)

		switch client.config.Discovery.Type {
		case string(discovery.LocalConfigServiceDiscoveryType):
			client.logger.Info("local discovery provider", zap.Reflect(`options`, client.config.Discovery.Options))

			client.discoveryProvider, err = discovery.NewLocalConfigProvider(client.config.Discovery.Options, mapper)
			if err != nil {
				return nil, fmt.Errorf(`initialize discovery provider: %w`, err)
			}

		case string(discovery.GossipServiceDiscoveryType):
			if client.config.Discovery.Connection == nil {
				return nil, ErrDiscoveryConnectionRequired
			}

			if client.discoverySigner == nil {
				return nil, ErrDiscoverySignerRequired
			}

			client.logger.Info("gossip discovery provider", zap.Reflect(`connection`, client.config.Discovery.Connection))

			identitySigner := func(msg []byte) ([]byte, error) {
				return client.discoverySigner.Sign(msg)
			}

			clientIdentity, err := client.discoverySigner.Serialize()
			if err != nil {
				return nil, fmt.Errorf(`serialize current defaultSigner: %w`, err)
			}

			// add tls settings from mapper if they were provided
			conn := mapper.MapConnection(client.config.Discovery.Connection.Host)
			client.config.Discovery.Connection.Tls = conn.TlsConfig
			client.config.Discovery.Connection.Host = conn.Host

			client.discoveryProvider, err = discovery.NewGossipDiscoveryProvider(
				client.ctx,
				*client.config.Discovery.Connection,
				client.logger,
				identitySigner,
				clientIdentity,
				mapper,
			)
			if err != nil {
				return nil, fmt.Errorf(`initialize discovery provider: %w`, err)
			}

			// discovery initialized, add local peers to the pool
			lDiscoverer, err := client.discoveryProvider.LocalPeers(client.ctx)
			if err != nil {
				return nil, fmt.Errorf(`fetch local peers from discovery provider connection=%s: %w`,
					client.config.Discovery.Connection.Host, err)
			}

			peers := lDiscoverer.Peers()

			for _, lp := range peers {
				mspID := lp.MspID

				for _, lpAddresses := range lp.HostAddresses {
					peerCfg := config.ConnectionConfig{
						Host: lpAddresses.Host,
						Tls:  lpAddresses.TlsConfig,
					}

					p, err := NewPeer(client.ctx, peerCfg, client.defaultSigner, client.logger)
					if err != nil {
						return nil, fmt.Errorf(`initialize endorsers for MSP: %s: %w`, mspID, err)
					}

					if err = client.peerPool.Add(mspID, p, StrategyGRPC(grpc.DefaultGrpcCheckPeriod)); err != nil {
						return nil, fmt.Errorf(`add peer to pool: %w`, err)
					}
				}
			}
		default:
			return nil, fmt.Errorf("unknown discovery type=%v. available: %v, %v",
				client.config.Discovery.Type,
				discovery.LocalConfigServiceDiscoveryType,
				discovery.GossipServiceDiscoveryType,
			)
		}
	}

	if client.orderer == nil && client.config != nil {
		client.logger.Info("initializing orderer")
		if len(client.config.Orderers) > 0 {
			ordConn, err := grpc.ConnectionFromConfigs(client.ctx, client.logger, client.config.Orderers...)
			if err != nil {
				return nil, fmt.Errorf(`initialize orderer connection: %w`, err)
			}

			client.orderer, err = NewOrdererFromGRPC(ordConn)
			if err != nil {
				return nil, fmt.Errorf(`initialize orderer: %w`, err)
			}
		}
	}

	return client, nil
}

func applyDefaults(c *Client) error {
	if c.logger == nil {
		c.logger = DefaultLogger
	}

	if c.crypto == nil {
		c.crypto = api.DefaultCryptoSuite()
	}

	return nil
}

func (c *Client) initPeerPool() error {
	c.logger.Info("initializing peer pool")
	if c.config.MSP == nil {
		return ErrEmptyMSPConfig
	}

	c.peerPool = NewPeerPool(c.ctx, c.logger)
	for _, mspConfig := range c.config.MSP {
		for _, peerConfig := range mspConfig.Endorsers {

			p, err := NewPeer(c.ctx, peerConfig, c.defaultSigner, c.logger)
			if err != nil {
				return fmt.Errorf("initialize endorsers for MSP: %s: %w", mspConfig.Name, err)
			}

			if err = c.peerPool.Add(mspConfig.Name, p, StrategyGRPC(grpc.DefaultGrpcCheckPeriod)); err != nil {
				return fmt.Errorf(`add peer to pool: %w`, err)
			}
		}
	}

	return nil
}

func (c *Client) CurrentIdentity() msp.SigningIdentity {
	return c.defaultSigner
}

func (c *Client) CryptoSuite() crypto.CryptoSuite {
	return c.crypto
}

func (c *Client) PeerPool() api.PeerPool {
	return c.peerPool
}

func (c *Client) FabricV2() bool {
	return c.fabricV2
}

func (c *Client) CurrentMspPeers() []api.Peer {
	allPeers := c.peerPool.GetPeers()

	if peers, ok := allPeers[c.defaultSigner.GetMSPIdentifier()]; !ok {
		return []api.Peer{}
	} else {
		return peers
	}
}

func (c *Client) Channel(name string) api.Channel {
	logger := c.logger.Named(`channel`).With(zap.String(`channel`, name))
	c.channelMx.Lock()
	defer c.channelMx.Unlock()

	ch, ok := c.channels[name]
	if ok {
		return ch
	}

	var ord api.Orderer

	logger.Debug(`channel instance doesn't exist, initiating new`)
	discChannel, err := c.discoveryProvider.Channel(c.ctx, name)
	if err != nil {
		logger.Error(`Failed channel discovery. We'll use default orderer`, zap.Error(err))
	} else {
		// if custom orderers are enabled
		if len(discChannel.Orderers()) > 0 {
			// convert api.HostEndpoint-> grpc config.ConnectionConfig
			var grpcConnCfgs []config.ConnectionConfig
			orderers := discChannel.Orderers()

			for _, orderer := range orderers {
				if len(orderer.HostAddresses) > 0 {
					for _, hostAddr := range orderer.HostAddresses {
						grpcCfg := config.ConnectionConfig{
							Host: hostAddr.Host,
							Tls:  hostAddr.TlsConfig,
						}
						grpcConnCfgs = append(grpcConnCfgs, grpcCfg)
					}
				}
			}
			// we can have many orderers and here we establish connection with internal round-robin balancer
			ordConn, err := grpc.ConnectionFromConfigs(c.ctx, c.logger, grpcConnCfgs...)
			if err != nil {
				logger.Error(`Failed to initialize custom GRPC connection for orderer`, zap.String(`channel`, name), zap.Error(err))
			}
			if ord, err = NewOrdererFromGRPC(ordConn); err != nil {
				logger.Error(`Failed to construct orderer from GRPC connection`)
			}
		}
	}

	// using default orderer
	if ord == nil {
		ord = c.orderer
	}

	ch = NewChannel(c.defaultSigner.GetMSPIdentifier(), name, c.peerPool, ord, c.discoveryProvider, c.defaultSigner, c.fabricV2, c.logger)
	c.channels[name] = ch
	return ch
}
