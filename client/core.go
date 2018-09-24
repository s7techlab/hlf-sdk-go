package client

import (
	"sync"
	"time"

	"github.com/s7techlab/hlf-sdk-go/peer/pool"

	"github.com/s7techlab/hlf-sdk-go/logger"

	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/client/chaincode/system"
	"github.com/s7techlab/hlf-sdk-go/client/channel"
	"github.com/s7techlab/hlf-sdk-go/crypto"
	"github.com/s7techlab/hlf-sdk-go/discovery"
	"github.com/s7techlab/hlf-sdk-go/orderer"
	"github.com/s7techlab/hlf-sdk-go/peer"
	"go.uber.org/zap"
)

type core struct {
	logger            *zap.Logger
	config            *config.Config
	mspId             string
	identity          msp.SigningIdentity
	peerPool          api.PeerPool
	localPeerDeliver  api.DeliverClient
	orderer           api.Orderer
	discoveryProvider api.DiscoveryProvider
	channels          map[string]api.Channel
	channelMx         sync.Mutex
	cs                api.CryptoSuite
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

func (c *core) Channel(name string) api.Channel {
	log := c.logger.Named(`Channel`).With(zap.String(`channel`, name))
	c.channelMx.Lock()
	defer c.channelMx.Unlock()
	log.Debug(`Check channel instance exists`)
	if ch, ok := c.channels[name]; ok {
		log.Debug(`Channel instance exists`)
		return ch
	} else {
		log.Debug(`Channel instance doesn't exist, initiating new`)
		ch = channel.NewCore(name, c.peerPool, c.orderer, c.discoveryProvider, c.identity, c.localPeerDeliver)
		c.channels[name] = ch
		return ch
	}
}

func NewCore(mspId string, identity api.Identity, opts ...CoreOpt) (api.Core, error) {
	var err error
	core := &core{
		mspId:    mspId,
		channels: make(map[string]api.Channel),
	}

	for _, option := range opts {
		if err = option(core); err != nil {
			return nil, errors.Wrap(err, `failed to apply option`)
		}
	}

	if core.config == nil {
		return nil, api.ErrEmptyConfig
	}

	if core.logger == nil {
		core.logger = logger.DefaultLogger
	}

	core.peerPool = pool.New(core.logger)

	if dp, err := discovery.GetProvider(core.config.Discovery.Type); err != nil {
		return nil, errors.Wrap(err, `failed to get discovery provider`)
	} else if core.discoveryProvider, err = dp.Initialize(core.config.Discovery.Options, core.peerPool); err != nil {
		return nil, errors.Wrap(err, `failed to initialize discovery provider`)
	}

	if core.cs, err = crypto.GetSuite(core.config.Crypto.Type, core.config.Crypto.Options); err != nil {
		return nil, errors.Wrap(err, `failed to initialize crypto suite`)
	}

	core.identity = identity.GetSigningIdentity(core.cs)

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

	if core.orderer == nil {
		if core.orderer, err = orderer.New(core.config.Orderer); err != nil {
			return nil, errors.Wrap(err, `failed to initialize orderer`)
		}
	}

	return core, nil
}
