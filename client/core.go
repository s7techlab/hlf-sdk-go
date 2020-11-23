package client

import (
	"context"
	"sync"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/platforms/golang"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/client/chaincode"
	"github.com/s7techlab/hlf-sdk-go/client/chaincode/system"
	"github.com/s7techlab/hlf-sdk-go/client/channel"
	"github.com/s7techlab/hlf-sdk-go/client/fetcher"
	"github.com/s7techlab/hlf-sdk-go/crypto"
	"github.com/s7techlab/hlf-sdk-go/discovery"
	"github.com/s7techlab/hlf-sdk-go/logger"
	"github.com/s7techlab/hlf-sdk-go/orderer"
	"github.com/s7techlab/hlf-sdk-go/peer"
	"github.com/s7techlab/hlf-sdk-go/peer/pool"
	"github.com/s7techlab/hlf-sdk-go/util"
)

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
	return system.NewSCC(c.peerPool, c.identity)
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
		discChannel, err := c.discoveryProvider.Channel(name)
		if err != nil {
			log.Error(`Failed to get channel declaration in discovery`, zap.Error(err))
		} else {
			// if custom orderers are enabled
			if len(discChannel.Orderers) > 0 {
				ordConn, err := util.NewGRPCConnectionFromConfigs(c.ctx, c.logger, discChannel.Orderers...)
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

		ch = channel.NewCore(c.mspId, name, c.peerPool, ord, c.discoveryProvider, c.identity, c.logger)
		c.channels[name] = ch
		return ch
	}
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

	if core.config == nil {
		return nil, api.ErrEmptyConfig
	}

	if core.ctx == nil {
		core.ctx = context.Background()
	}

	if core.logger == nil {
		core.logger = logger.DefaultLogger
	}

	if dp, err := discovery.GetProvider(core.config.Discovery.Type); err != nil {
		return nil, errors.Wrap(err, `failed to get discovery provider`)
	} else if core.discoveryProvider, err = dp.Initialize(core.config.Discovery.Options, core.peerPool); err != nil {
		return nil, errors.Wrap(err, `failed to initialize discovery provider`)
	}

	if core.cs, err = crypto.GetSuite(core.config.Crypto.Type, core.config.Crypto.Options); err != nil {
		return nil, errors.Wrap(err, `failed to initialize crypto suite`)
	}

	core.identity = identity.GetSigningIdentity(core.cs)

	// if peerPool is empty, set it from config
	if core.peerPool == nil {
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

	if core.orderer == nil {
		if core.orderer, err = orderer.New(core.config.Orderer, core.logger); err != nil {
			return nil, errors.Wrap(err, `failed to initialize orderer`)
		}
	}

	// use chaincode fetcher for Go chaincodes by default
	if core.fetcher == nil {
		core.fetcher = fetcher.NewLocal(&golang.Platform{})
	}

	return core, nil
}
