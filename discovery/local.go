package discovery

import (
	"context"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
)

// implementation of api.DiscoveryProvider interface
var _ api.DiscoveryProvider = (*LocalConfigDiscoveryProvider)(nil)

type LocalConfigDiscoveryProvider struct {
	opts opts
}

type opts struct {
	Channels []api.DiscoveryChannel `yaml:"channels"`
}

func (d *LocalConfigDiscoveryProvider) Chaincode(_ context.Context, channelName, ccName string) (api.ChaincodeDiscoverer, error) {
	var channelFoundFlag bool

	for _, ch := range d.opts.Channels {
		if ch.Name == channelName {
			channelFoundFlag = true
			for _, cc := range ch.Chaincodes {
				if cc.Name == ccName {
					// TODO from where to get endorsers
					// no endorsers in local cfg
					// no peers
					// endorsers := []*api.HostEndpoint{}

					dc := newChaincodeDTO(cc.Name, cc.Version, channelName)
					for i := range ch.Orderers {
						mspID := "" // TODO we have no MSPID from local cfg
						dc.addEndpointToOrderers(mspID, ch.Orderers[i].Host)
					}

					return dc, nil
				}
			}
		}
	}

	if channelFoundFlag {
		return nil, ErrNoChaincodes
	}
	return nil, ErrChannelNotFound
}

func (d *LocalConfigDiscoveryProvider) Channel(_ context.Context, channelName string) (api.ChannelDiscoverer, error) {
	var channelFoundFlag bool

	for _, ch := range d.opts.Channels {
		if ch.Name == channelName {
			channelFoundFlag = true

			dc := newChannelDTO(channelName)
			for i := range ch.Orderers {
				mspID := "" // TODO we have no MSPID from local cfg
				dc.addEndpointToOrderers(mspID, ch.Orderers[i].Host)
			}

			return dc, nil
		}
	}

	if channelFoundFlag {
		return nil, ErrNoChaincodes
	}
	return nil, ErrChannelNotFound
}

func NewLocalConfigDiscoveryProvider(options config.DiscoveryConfigOpts) (api.DiscoveryProvider, error) {
	var opts opts
	if err := mapstructure.Decode(options, &opts); err != nil {
		return nil, errors.Wrap(err, `failed to decode params`)
	}

	return &LocalConfigDiscoveryProvider{opts: opts}, nil
}
