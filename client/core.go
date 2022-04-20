package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/client/chaincode"
	"github.com/s7techlab/hlf-sdk-go/client/chaincode/system"
	"github.com/s7techlab/hlf-sdk-go/client/channel"
	"github.com/s7techlab/hlf-sdk-go/crypto"
	"github.com/s7techlab/hlf-sdk-go/crypto/ecdsa"
	"github.com/s7techlab/hlf-sdk-go/discovery"
	"github.com/s7techlab/hlf-sdk-go/orderer"
	"github.com/s7techlab/hlf-sdk-go/peer"
	"github.com/s7techlab/hlf-sdk-go/peer/pool"
	"github.com/s7techlab/hlf-sdk-go/util"
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
	chaincodes        map[string]api.ChaincodePackage
	chaincodeMx       sync.Mutex
	cs                api.CryptoSuite
	fetcher           api.CCFetcher
	fabricV2          bool
}

func (c *core) ChaincodeLifecycle() api.Lifecycle {
	return system.NewLifecycle(c)
}

func (c *core) Chaincode(name string) api.ChaincodePackage {
	c.chaincodeMx.Lock()
	defer c.chaincodeMx.Unlock()

	cc, ok := c.chaincodes[name]
	if !ok {
		cc = chaincode.NewCorePackage(name, system.NewLSCC(c.peerPool, c.identity), c.fetcher, c.orderer, c.identity)
		c.chaincodes[name] = cc
		return cc
	}

	return cc
}

func (c *core) System() api.SystemCC {
	return system.NewSCC(c)
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

func (c *core) CurrentMspPeers() []api.Peer {
	allPeers := c.peerPool.GetPeers()

	if peers, ok := allPeers[c.identity.GetMSPIdentifier()]; !ok {
		return []api.Peer{}
	} else {
		return peers
	}
}

func (c *core) Channel(name string) api.Channel {
	log := c.logger.Named(`channel`).With(zap.String(`channel`, name))
	c.channelMx.Lock()
	defer c.channelMx.Unlock()

	ch, ok := c.channels[name]
	if ok {
		return ch
	}

	var ord api.Orderer

	log.Debug(`Channel instance doesn't exist, initiating new`)
	discChannel, err := c.discoveryProvider.Channel(c.ctx, name)
	if err != nil {
		log.Error(`Failed channel discovery. We'll use default orderer`, zap.Error(err))
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
							Host: hostAddr.Address,
							Tls:  hostAddr.TLSSettings,
						}
						grpcConnCfgs = append(grpcConnCfgs, grpcCfg)
					}
				}
			}
			// we can have many orderers and here we establish connection with internal round-robin balancer
			ordConn, err := util.NewGRPCConnectionFromConfigs(c.ctx, c.logger, grpcConnCfgs...)
			if err != nil {
				log.Error(`Failed to initialize custom GRPC connection for orderer`, zap.String(`channel`, name), zap.Error(err))
			}
			if ord, err = orderer.NewFromGRPC(c.ctx, ordConn); err != nil {
				log.Error(`Failed to construct orderer from GRPC connection`)
			}
		}
	}

	// using default orderer
	if ord == nil {
		ord = c.orderer
	}

	ch = channel.NewCore(c.identity.GetMSPIdentifier(), name, c.peerPool, ord, c.discoveryProvider, c.identity, c.fabricV2, c.logger)
	c.channels[name] = ch
	return ch
}

func (c *core) FabricV2() bool {
	return c.fabricV2
}

// Deprecated: use New
func NewCore(identity api.Identity, opts ...CoreOpt) (api.Core, error) {
	return New(identity, opts...)
}

func New(identity api.Identity, opts ...CoreOpt) (api.Core, error) {
	var err error
	core := &core{
		channels:   make(map[string]api.Channel),
		chaincodes: make(map[string]api.ChaincodePackage),
	}

	for _, option := range opts {
		if err = option(core); err != nil {
			return nil, fmt.Errorf(`apply option: %w`, err)
		}
	}

	if core.ctx == nil {
		core.ctx = context.Background()
	}

	if core.logger == nil {
		core.logger = DefaultLogger
	}

	if core.cs == nil {

		if core.config == nil {
			return nil, api.ErrEmptyConfig
		}
		if core.config.Crypto.Type == `` {
			core.logger.Debug("crypto suite: use default config")
			core.config.Crypto = ecdsa.DefaultConfig
		}
		if core.cs, err = crypto.GetSuite(core.config.Crypto.Type, core.config.Crypto.Options); err != nil {
			return nil, fmt.Errorf(`initialize crypto suite: %w`, err)
		}
	}

	if identity == nil {
		return nil, errors.New("identity wasn't provided")
	}

	core.identity = identity.GetSigningIdentity(core.cs)

	// if peerPool is empty, set it from config
	if core.peerPool == nil {
		core.logger.Info("initializing peer pool")

		if core.config == nil {
			return nil, api.ErrEmptyConfig
		}
		core.peerPool = pool.New(core.ctx, core.logger)
		for _, mspConfig := range core.config.MSP {
			for _, peerConfig := range mspConfig.Endorsers {
				var p api.Peer
				p, err = peer.New(peerConfig, core.logger)
				if err != nil {
					return nil, fmt.Errorf("initialize endorsers for MSP: %s: %w", mspConfig.Name, err)
				}
				if err = core.peerPool.Add(mspConfig.Name, p, api.StrategyGRPC(5*time.Second)); err != nil {
					return nil, fmt.Errorf(`add peer to pool: %w`, err)
				}
			}
		}
	}

	if core.discoveryProvider == nil && core.config != nil {

		tlsMapper := discovery.NewTLSCertsMapper(core.config.TLSCertsMap)

		switch core.config.Discovery.Type {
		case string(discovery.LocalConfigServiceDiscoveryType):
			core.logger.Info("local discovery provider", zap.Reflect(`options`, core.config.Discovery.Options))
			core.discoveryProvider, err = discovery.NewLocalConfigProvider(core.config.Discovery.Options, tlsMapper)
			if err != nil {
				return nil, fmt.Errorf(`initialize discovery provider: %w`, err)
			}
		case string(discovery.GossipServiceDiscoveryType):
			if core.config.Discovery.Connection == nil {
				return nil, fmt.Errorf(`discovery connection config wasn't provided. configure 'discovery.connection': %w`, err)
			}
			core.logger.Info("gossip discovery provider", zap.Reflect(`connection`, core.config.Discovery.Connection))

			identitySigner := func(msg []byte) ([]byte, error) {
				return core.CurrentIdentity().Sign(msg)
			}
			clientIdentity, err := core.CurrentIdentity().Serialize()
			if err != nil {
				return nil, fmt.Errorf(`serialize current identity: %w`, err)
			}
			// add tls settings from mapper if they were provided
			core.config.Discovery.Connection.Tls = *tlsMapper.TlsConfigForAddress(core.config.Discovery.Connection.Host)

			core.discoveryProvider, err = discovery.NewGossipDiscoveryProvider(
				core.ctx,
				*core.config.Discovery.Connection,
				core.logger,
				identitySigner,
				clientIdentity,
				tlsMapper,
			)
			if err != nil {
				return nil, fmt.Errorf(`initialize discovery provider: %w`, err)
			}
			// discovery initialized, add local peers to the pool
			lDiscoverer, err := core.discoveryProvider.LocalPeers(core.ctx)
			if err != nil {
				return nil, fmt.Errorf(`fetch local peers from discovery provider connection=%s: %w`,
					core.config.Discovery.Connection.Host, err)
			}

			peers := lDiscoverer.Peers()

			for _, lp := range peers {
				mspID := lp.MspID

				for _, lpAddresses := range lp.HostAddresses {
					peerCfg := config.ConnectionConfig{
						Host: lpAddresses.Address,
						Tls:  lpAddresses.TLSSettings,
					}
					p, err := peer.New(peerCfg, core.logger)
					if err != nil {
						return nil, fmt.Errorf(`initialize endorsers for MSP: %s: %w`, mspID, err)
					}
					if err = core.peerPool.Add(mspID, p, api.StrategyGRPC(5*time.Second)); err != nil {
						return nil, fmt.Errorf(`add peer to pool: %w`, err)
					}
				}
			}
		default:
			return nil, fmt.Errorf("unknown discovery type=%v. available: %v, %v",
				core.config.Discovery.Type,
				discovery.LocalConfigServiceDiscoveryType,
				discovery.GossipServiceDiscoveryType,
			)
		}
	}

	if core.orderer == nil && core.config != nil {
		core.logger.Info("initializing orderer")
		if len(core.config.Orderers) > 0 {
			ordConn, err := util.NewGRPCConnectionFromConfigs(core.ctx, core.logger, core.config.Orderers...)
			if err != nil {
				return nil, fmt.Errorf(`initialize orderer connection: %w`, err)
			}
			core.orderer, err = orderer.NewFromGRPC(core.ctx, ordConn)
			if err != nil {
				return nil, fmt.Errorf(`initialize orderer: %w`, err)
			}
		}
	}

	//// use chaincode fetcher for Go chaincodes by default
	//if core.fetcher == nil {
	//	core.fetcher = fetcher.NewLocal(&golang.Platform{})
	//}

	return core, nil
}
