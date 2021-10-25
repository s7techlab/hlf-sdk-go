package client

import (
	"context"
	"sync"
	"time"

	fabOrderer "github.com/hyperledger/fabric-protos-go/orderer"
	fabPeer "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/core/chaincode/platforms/golang"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/v2/api"
	"github.com/s7techlab/hlf-sdk-go/v2/api/config"
	"github.com/s7techlab/hlf-sdk-go/v2/client/chaincode"
	"github.com/s7techlab/hlf-sdk-go/v2/client/chaincode/system"
	"github.com/s7techlab/hlf-sdk-go/v2/client/channel"
	"github.com/s7techlab/hlf-sdk-go/v2/client/fetcher"
	"github.com/s7techlab/hlf-sdk-go/v2/crypto"
	"github.com/s7techlab/hlf-sdk-go/v2/crypto/ecdsa"
	"github.com/s7techlab/hlf-sdk-go/v2/discovery"
	"github.com/s7techlab/hlf-sdk-go/v2/logger"
	"github.com/s7techlab/hlf-sdk-go/v2/orderer"
	"github.com/s7techlab/hlf-sdk-go/v2/peer"
	"github.com/s7techlab/hlf-sdk-go/v2/peer/pool"
	"github.com/s7techlab/hlf-sdk-go/v2/util"
)

// implementation of api.DiscoveryProvider interface
var _ api.Core = (*core)(nil)

type core struct {
	ctx               context.Context
	logger            *zap.Logger
	config            *config.Config
	mspId             string
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

func (c *core) Chaincode(name string) api.ChaincodePackage {
	c.chaincodeMx.Lock()
	defer c.chaincodeMx.Unlock()
	if cc, ok := c.chaincodes[name]; !ok {
		cc = chaincode.NewCorePackage(name, system.NewLSCC(c.peerPool, c.identity), c.fetcher, c.orderer, c.identity)
		c.chaincodes[name] = cc
		return cc
	} else {
		return cc
	}
}

func (c *core) System() api.SystemCC {
	return system.NewSCC(c.peerPool, c.identity, c.fabricV2)
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

func (c *core) Channel(name string) api.Channel {
	log := c.logger.Named(`Channel`).With(zap.String(`channel`, name))
	c.channelMx.Lock()
	defer c.channelMx.Unlock()
	if ch, ok := c.channels[name]; ok {
		return ch
	} else {
		var ord api.Orderer

		log.Debug(`Channel instance doesn't exist, initiating new`)
		discChannel, err := c.discoveryProvider.Channel(c.ctx, name)
		if err != nil {
			log.Error(`Failed to get channel declaration in discovery`, zap.Error(err))
		} else {
			// if custom orderers are enabled
			if len(discChannel.Orderers()) > 0 {
				// convert api.HostEndpoint-> grpc config.ConnectionConfig
				grpcConnCfgs := []config.ConnectionConfig{}
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

		ch = channel.NewCore(c.mspId, name, c.peerPool, ord,
			c.discoveryProvider, c.identity, c.fabricV2, c.logger)
		c.channels[name] = ch
		return ch
	}
}

func (c *core) FabricV2() bool {
	return c.fabricV2
}

func (c *core) ChannelChaincode(ctx context.Context, chanName string, ccName string) (api.Chaincode, error) {
	return c.Channel(chanName).Chaincode(ctx, ccName)
}

func (c *core) Events(
	ctx context.Context,
	chanName string,
	ccName string,
	eventCCSeekOption ...func() (*fabOrderer.SeekPosition, *fabOrderer.SeekPosition),
) (chan *fabPeer.ChaincodeEvent, error) {
	identity := c.CurrentIdentity()
	mspID := identity.GetMSPIdentifier()

	dc, err := c.PeerPool().DeliverClient(mspID, identity)
	if err != nil {
		return nil, err
	}

	seekOpts := []api.EventCCSeekOption{}
	for i := range eventCCSeekOption {
		seekOpts = append(seekOpts, eventCCSeekOption[i])
	}

	subcription, err := dc.SubscribeCC(ctx, chanName, ccName, seekOpts...)
	if err != nil {
		return nil, err
	}

	return subcription.Events(), nil
}

func NewCore(mspId string, identity api.Identity, opts ...CoreOpt) (api.Core, error) {
	var err error
	core := &core{
		mspId:      mspId,
		channels:   make(map[string]api.Channel),
		chaincodes: make(map[string]api.ChaincodePackage),
	}

	for _, option := range opts {
		if err = option(core); err != nil {
			return nil, errors.Wrap(err, `failed to apply option`)
		}
	}

	if core.ctx == nil {
		core.ctx = context.Background()
	}

	if core.logger == nil {
		core.logger = logger.DefaultLogger
	}

	if core.cs == nil {
		core.logger.Info("initializing crypto suite")

		if core.config == nil {
			return nil, api.ErrEmptyConfig
		}
		if core.config.Crypto.Type == `` {
			core.config.Crypto = ecdsa.DefaultConfig
		}
		if core.cs, err = crypto.GetSuite(core.config.Crypto.Type, core.config.Crypto.Options); err != nil {
			return nil, errors.Wrap(err, `failed to initialize crypto suite`)
		}
	}

	core.identity = identity.GetSigningIdentity(core.cs)

	// if peerPool is empty, set it from config
	if core.peerPool == nil {
		core.logger.Info("initializing peer pool")

		if core.config == nil {
			return nil, api.ErrEmptyConfig
		}
		core.peerPool = pool.New(core.ctx, core.logger, core.config.Pool)
		for _, mspConfig := range core.config.MSP {
			for _, peerConfig := range mspConfig.Endorsers {
				if p, err := peer.New(peerConfig, core.logger); err != nil {
					return nil, errors.Errorf("failed to initialize endorsers for MSP: %s:%s", mspConfig.Name, err.Error())
				} else {
					if err = core.peerPool.Add(mspConfig.Name, p, api.StrategyGRPC(5*time.Second)); err != nil {
						return nil, errors.Wrap(err, `failed to add peer to pool`)
					}
				}
			}
		}
	}

	if core.discoveryProvider == nil && core.config != nil {
		core.logger.Info("initializing discovery provider")

		tlsMapper := discovery.NewTLSCertsMapper(core.config.TLSCertsMap)

		switch core.config.Discovery.Type {
		case string(discovery.LocalConfigServiceDiscoveryType):
			core.discoveryProvider, err = discovery.NewLocalConfigProvider(core.config.Discovery.Options, tlsMapper)
			if err != nil {
				return nil, errors.Wrap(err, `failed to initialize discovery provider`)
			}
		case string(discovery.GossipServiceDiscoveryType):
			if core.config.Discovery.Connection == nil {
				return nil, errors.Wrap(err, `discovery connection config wasn't provided. configure 'discovery.connection'`)
			}
			identitySigner := func(msg []byte) ([]byte, error) {
				return core.CurrentIdentity().Sign(msg)
			}
			clientIdentity, err := core.CurrentIdentity().Serialize()
			if err != nil {
				return nil, errors.Wrap(err, `failed serialize current identity`)
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
				return nil, errors.Wrap(err, `failed to initialize discovery provider`)
			}
			// discovery initialized, add local peers to the pool
			lDiscoverer, err := core.discoveryProvider.LocalPeers(core.ctx)
			if err != nil {
				return nil, errors.Wrap(err, `failed to fetch local peers`)
			}

			peers := lDiscoverer.Peers()

			for _, lp := range peers {
				mspID := lp.MspID

				for _, lpAddresses := range lp.HostAddresses {
					peerCfg := config.ConnectionConfig{
						Host: lpAddresses.Address,
						Tls:  lpAddresses.TLSSettings,
					}

					if p, err := peer.New(peerCfg, core.logger); err != nil {
						return nil, errors.Errorf("failed to initialize endorsers for MSP: %s:%s", mspID, err.Error())
					} else {
						if err = core.peerPool.Add(mspID, p, api.StrategyGRPC(5*time.Second)); err != nil {
							return nil, errors.Wrap(err, `failed to add peer to pool`)
						}
					}
				}
			}
		}
	}

	if core.orderer == nil && core.config != nil {
		core.logger.Info("initializing orderer")
		if len(core.config.Orderers) > 0 {
			ordConn, err := util.NewGRPCConnectionFromConfigs(core.ctx, core.logger, core.config.Orderers...)
			if err != nil {
				return nil, errors.Wrap(err, `failed to initialize orderer connection`)
			}
			core.orderer, err = orderer.NewFromGRPC(core.ctx, ordConn)
			if err != nil {
				return nil, errors.Wrap(err, `failed to initialize orderer`)
			}
		} else if core.config.Orderer != nil {
			core.orderer, err = orderer.New(*core.config.Orderer, core.logger)
			if err != nil {
				return nil, errors.Wrap(err, `failed to initialize orderer`)
			}
		}
	}

	// use chaincode fetcher for Go chaincodes by default
	if core.fetcher == nil {
		core.fetcher = fetcher.NewLocal(&golang.Platform{})
	}

	return core, nil
}
