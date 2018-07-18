package local

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/discovery"
)

const Name = `local`

type discoveryProvider struct {
	opts opts
}

type opts struct {
	Channels []api.DiscoveryChannel `yaml:"channels"`
}

func (d *discoveryProvider) Channels() ([]api.DiscoveryChannel, error) {
	if len(d.opts.Channels) > 0 {
		return d.opts.Channels, nil
	}
	return nil, discovery.ErrNoChannels
}

func (d *discoveryProvider) Chaincode(channelName string, ccName string) (*api.DiscoveryChaincode, error) {
	var channelFoundFlag bool

	for _, ch := range d.opts.Channels {
		if ch.Name == channelName {
			channelFoundFlag = true
			for _, cc := range ch.Chaincodes {
				if cc.Name == ccName {
					return &cc, nil
				}
			}
		}
	}

	if channelFoundFlag {
		return nil, discovery.ErrNoChaincodes
	}
	return nil, discovery.ErrChannelNotFound
}

func (d *discoveryProvider) Chaincodes(channelName string) ([]api.DiscoveryChaincode, error) {
	for _, ch := range d.opts.Channels {
		if ch.Name == channelName {
			return ch.Chaincodes, nil
		}
	}
	return nil, discovery.ErrChannelNotFound
}

func (d *discoveryProvider) Endorsers(channelName string, ccName string) ([]config.PeerConfig, error) {

	var channelFoundFlag bool

	for _, ch := range d.opts.Channels {
		if ch.Name == channelName {
			channelFoundFlag = true
			for _, cc := range ch.Chaincodes {
				if cc.Name == ccName {
					if len(ch.Peers) > 0 {
						return ch.Peers, nil
					} else {
						return nil, discovery.ErrNoEndorsers
					}
				}
			}
		}
	}

	if channelFoundFlag {
		return nil, discovery.ErrNoChannels
	}
	return nil, discovery.ErrChannelNotFound
}

func (d *discoveryProvider) Initialize(options config.DiscoveryConfigOpts) (api.DiscoveryProvider, error) {
	var opts opts
	if err := mapstructure.Decode(options, &opts); err != nil {
		return nil, errors.Wrap(err, `failed to decode params`)
	}

	return &discoveryProvider{opts: opts}, nil
}

func init() {
	discovery.SetProvider(Name, &discoveryProvider{})
}
