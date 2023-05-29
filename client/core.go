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
	"github.com/s7techlab/hlf-sdk-go/client/grpc"
	"github.com/s7techlab/hlf-sdk-go/crypto"
	"github.com/s7techlab/hlf-sdk-go/crypto/ecdsa"
	"github.com/s7techlab/hlf-sdk-go/discovery"
)

// implementation of api.Core interface
var _ api.Core = (*core)(nil)

type core struct {
	ctx               context.Context
	logger            *zap.Logger
	config            *config.Config
	identity          msp.SigningIdentity
	peerPool          api.PeerPool
	orderer           api.Orderer
	discoveryProvider api.DiscoveryProvider
	channels          map[string]api.Channel
	channelMx         sync.Mutex
	cs                api.CryptoSuite
	fabricV2          bool
}

func New(identity api.Identity, opts ...CoreOpt) (api.Core, error) {
	if identity == nil {
		return nil, errors.New("identity wasn't provided")
	}

	coreImpl := &core{
		channels: make(map[string]api.Channel),
	}

	for _, option := range opts {
		if err := option(coreImpl); err != nil {
			return nil, fmt.Errorf(`apply option: %w`, err)
		}
	}

	if coreImpl.ctx == nil {
		coreImpl.ctx = context.Background()
	}

	if coreImpl.logger == nil {
		coreImpl.logger = DefaultLogger
	}

	var err error
	if coreImpl.cs == nil {
		coreImpl.cs, err = crypto.GetSuite(ecdsa.DefaultConfig.Type, ecdsa.DefaultConfig.Options)
		if err != nil {
			return nil, fmt.Errorf(`initialize crypto suite: %w`, err)
		}
	}

	coreImpl.identity = identity.GetSigningIdentity(coreImpl.cs)

	// if peerPool is empty, set it from config
	if coreImpl.peerPool == nil {
		coreImpl.logger.Info("initializing peer pool")

		if coreImpl.config == nil {
			return nil, api.ErrEmptyConfig
		}

		coreImpl.peerPool = NewPeerPool(coreImpl.ctx, coreImpl.logger)
		for _, mspConfig := range coreImpl.config.MSP {
			for _, peerConfig := range mspConfig.Endorsers {
				var p api.Peer
				p, err = NewPeer(coreImpl.ctx, peerConfig, coreImpl.identity, coreImpl.logger)
				if err != nil {
					return nil, fmt.Errorf("initialize endorsers for MSP: %s: %w", mspConfig.Name, err)
				}

				if err = coreImpl.peerPool.Add(mspConfig.Name, p, api.StrategyGRPC(api.DefaultGrpcCheckPeriod)); err != nil {
					return nil, fmt.Errorf(`add peer to pool: %w`, err)
				}
			}
		}
	}

	if coreImpl.discoveryProvider == nil && coreImpl.config != nil {
		mapper := discovery.NewEndpointsMapper(coreImpl.config.EndpointsMap)

		switch coreImpl.config.Discovery.Type {
		case string(discovery.LocalConfigServiceDiscoveryType):
			coreImpl.logger.Info("local discovery provider", zap.Reflect(`options`, coreImpl.config.Discovery.Options))

			coreImpl.discoveryProvider, err = discovery.NewLocalConfigProvider(coreImpl.config.Discovery.Options, mapper)
			if err != nil {
				return nil, fmt.Errorf(`initialize discovery provider: %w`, err)
			}

		case string(discovery.GossipServiceDiscoveryType):
			if coreImpl.config.Discovery.Connection == nil {
				return nil, fmt.Errorf(`discovery connection config wasn't provided. configure 'discovery.connection': %w`, err)
			}

			coreImpl.logger.Info("gossip discovery provider", zap.Reflect(`connection`, coreImpl.config.Discovery.Connection))

			identitySigner := func(msg []byte) ([]byte, error) {
				return coreImpl.CurrentIdentity().Sign(msg)
			}

			clientIdentity, err := coreImpl.CurrentIdentity().Serialize()
			if err != nil {
				return nil, fmt.Errorf(`serialize current identity: %w`, err)
			}

			// add tls settings from mapper if they were provided
			conn := mapper.MapConnection(coreImpl.config.Discovery.Connection.Host)
			coreImpl.config.Discovery.Connection.Tls = conn.TlsConfig
			coreImpl.config.Discovery.Connection.Host = conn.Host

			coreImpl.discoveryProvider, err = discovery.NewGossipDiscoveryProvider(
				coreImpl.ctx,
				*coreImpl.config.Discovery.Connection,
				coreImpl.logger,
				identitySigner,
				clientIdentity,
				mapper,
			)
			if err != nil {
				return nil, fmt.Errorf(`initialize discovery provider: %w`, err)
			}

			// discovery initialized, add local peers to the pool
			lDiscoverer, err := coreImpl.discoveryProvider.LocalPeers(coreImpl.ctx)
			if err != nil {
				return nil, fmt.Errorf(`fetch local peers from discovery provider connection=%s: %w`,
					coreImpl.config.Discovery.Connection.Host, err)
			}

			peers := lDiscoverer.Peers()

			for _, lp := range peers {
				mspID := lp.MspID

				for _, lpAddresses := range lp.HostAddresses {
					peerCfg := config.ConnectionConfig{
						Host: lpAddresses.Host,
						Tls:  lpAddresses.TlsConfig,
					}

					p, err := NewPeer(coreImpl.ctx, peerCfg, coreImpl.identity, coreImpl.logger)
					if err != nil {
						return nil, fmt.Errorf(`initialize endorsers for MSP: %s: %w`, mspID, err)
					}

					if err = coreImpl.peerPool.Add(mspID, p, api.StrategyGRPC(api.DefaultGrpcCheckPeriod)); err != nil {
						return nil, fmt.Errorf(`add peer to pool: %w`, err)
					}
				}
			}
		default:
			return nil, fmt.Errorf("unknown discovery type=%v. available: %v, %v",
				coreImpl.config.Discovery.Type,
				discovery.LocalConfigServiceDiscoveryType,
				discovery.GossipServiceDiscoveryType,
			)
		}
	}

	if coreImpl.orderer == nil && coreImpl.config != nil {
		coreImpl.logger.Info("initializing orderer")
		if len(coreImpl.config.Orderers) > 0 {
			ordConn, err := grpc.ConnectionFromConfigs(coreImpl.ctx, coreImpl.logger, coreImpl.config.Orderers...)
			if err != nil {
				return nil, fmt.Errorf(`initialize orderer connection: %w`, err)
			}

			coreImpl.orderer, err = NewOrdererFromGRPC(ordConn)
			if err != nil {
				return nil, fmt.Errorf(`initialize orderer: %w`, err)
			}
		}
	}

	return coreImpl, nil
}

func (c *core) CurrentIdentity() msp.SigningIdentity {
	return c.identity
}

func (c *core) CryptoSuite() api.CryptoSuite {
	return c.cs
}

func (c *core) PeerPool() api.PeerPool {
	return c.peerPool
}

func (c *core) FabricV2() bool {
	return c.fabricV2
}

func (c *core) CurrentMspPeers() []api.Peer {
	allPeers := c.peerPool.GetPeers()

	if peers, ok := allPeers[c.identity.GetMSPIdentifier()]; !ok {
		return []api.Peer{}
	} else {
		return peers
	}
}

func (c *core) Channel(name string) api.Channel {
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

	ch = NewChannel(c.identity.GetMSPIdentifier(), name, c.peerPool, ord, c.discoveryProvider, c.identity, c.fabricV2, c.logger)
	c.channels[name] = ch
	return ch
}
