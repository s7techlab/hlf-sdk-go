package discovery

import (
	"context"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/discovery"
)

type LocalConfigDiscoveryProvider struct {
	opts opts
}

type opts struct {
	Channels []api.DiscoveryChannel `yaml:"channels"`
}

func (d *LocalConfigDiscoveryProvider) Chaincode(_ context.Context, channelName string, ccName string) (api.IDiscoveryChaincode, error) {
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

					dc := newDiscoveryChaincode(cc.Name, cc.Version, channelName)
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
		return nil, discovery.ErrNoChaincodes
	}
	return nil, discovery.ErrChannelNotFound
}

func NewLocalConfigDiscoveryProvider(options config.DiscoveryConfigOpts) (api.DiscoveryProvider, error) {
	var opts opts
	if err := mapstructure.Decode(options, &opts); err != nil {
		return nil, errors.Wrap(err, `failed to decode params`)
	}

	return &LocalConfigDiscoveryProvider{opts: opts}, nil
}
