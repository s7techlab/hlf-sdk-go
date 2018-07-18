package member

import (
	"sync"

	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/crypto"
	"github.com/s7techlab/hlf-sdk-go/discovery"
	"github.com/s7techlab/hlf-sdk-go/member/chaincode/system"
	"github.com/s7techlab/hlf-sdk-go/member/channel"
	"github.com/s7techlab/hlf-sdk-go/orderer"
	"github.com/s7techlab/hlf-sdk-go/peer"
)

type Core struct {
	mspId             string
	identity          msp.SigningIdentity
	localPeer         api.Peer
	localEventHub     api.EventHub
	orderer           api.Orderer
	discoveryProvider api.DiscoveryProvider
	channels          map[string]api.Channel
	channelMx         sync.Mutex
	cs                api.CryptoSuite
}

func (c *Core) System() api.SystemCC {
	return system.NewSCC(c.localPeer, c.identity)
}

func (c *Core) CurrentIdentity() msp.SigningIdentity {
	return c.identity
}

func (c *Core) CryptoSuite() api.CryptoSuite {
	return c.cs
}

func (c *Core) Channel(name string) api.Channel {
	c.channelMx.Lock()
	defer c.channelMx.Unlock()
	if ch, ok := c.channels[name]; ok {
		return ch
	} else {
		ch = channel.NewCore(name, c.localPeer, c.orderer, c.discoveryProvider, c.identity, c.localEventHub)
		c.channels[name] = ch
		return ch
	}
}

func NewCore(mspId string, configPath string, identity api.Identity) (*Core, error) {
	conf, err := config.NewYamlConfig(configPath)
	if err != nil {
		return nil, errors.Wrap(err, `failed to initialize config`)
	}

	core := Core{
		mspId:    mspId,
		channels: make(map[string]api.Channel),
	}

	if dp, err := discovery.GetProvider(conf.Discovery.Type); err != nil {
		return nil, errors.Wrap(err, `failed to get discovery provider`)
	} else if core.discoveryProvider, err = dp.Initialize(conf.Discovery.Options); err != nil {
		return nil, errors.Wrap(err, `failed to initialize discovery provider`)
	}

	if core.cs, err = crypto.GetSuite(conf.Crypto.Type, conf.Crypto.Options); err != nil {
		return nil, errors.Wrap(err, `failed to initialize crypto suite`)
	}

	if core.localPeer, err = peer.New(conf.LocalPeer); err != nil {
		return nil, errors.Wrap(err, `failed to initialize local peer`)
	}

	if core.orderer, err = orderer.New(conf.Orderer); err != nil {
		return nil, errors.Wrap(err, `failed to initialize orderer`)
	}

	core.identity = identity.GetSigningIdentity(core.cs)

	if core.localEventHub, err = peer.NewEventHub(conf.LocalPeer, core.identity); err != nil {
		return nil, errors.Wrap(err, `failed to initialize event hub`)
	}

	return &core, nil
}
