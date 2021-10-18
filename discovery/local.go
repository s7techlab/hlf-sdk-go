package discovery

import (
	"context"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
)

// implementation of api.DiscoveryProvider interface
var _ api.DiscoveryProvider = (*LocalConfigProvider)(nil)

type LocalConfigProvider struct {
	tlsMapper tlsConfigMapper
	channels  []config.DiscoveryChannel `yaml:"channels"`
}

func (d *LocalConfigProvider) Chaincode(_ context.Context, channelName, ccName string) (api.ChaincodeDiscoverer, error) {
	var channelFoundFlag bool

	for _, ch := range d.channels {
		if ch.Name == channelName {
			channelFoundFlag = true
			for _, cc := range ch.Chaincodes {
				if cc.Name == ccName {
					// TODO from where to get endorsers
					// no endorsers in local cfg
					// no peers
					// endorsers := []*api.HostEndpoint{}

					ccDTO := newChaincodeDTO(cc.Name, cc.Version, channelName)
					for i := range ch.Orderers {
						mspID := "" // TODO we have no MSPID from local cfg
						ccDTO.addEndpointToOrderers(mspID, ch.Orderers[i].Host)
					}

					return newChaincodeDiscovererTLSDecorator(ccDTO, d.tlsMapper), nil
				}
			}
		}
	}

	if channelFoundFlag {
		return nil, ErrNoChaincodes
	}
	return nil, ErrChannelNotFound
}

func (d *LocalConfigProvider) Channel(_ context.Context, channelName string) (api.ChannelDiscoverer, error) {
	var channelFoundFlag bool

	for _, ch := range d.channels {
		if ch.Name == channelName {
			channelFoundFlag = true

			chanDTO := newChannelDTO(channelName)
			for i := range ch.Orderers {
				mspID := "" // TODO we have no MSPID from local cfg
				chanDTO.addEndpointToOrderers(mspID, ch.Orderers[i].Host)
			}

			return newChannelDiscovererTLSDecorator(chanDTO, d.tlsMapper), nil
		}
	}

	if channelFoundFlag {
		return nil, ErrNoChaincodes
	}
	return nil, ErrChannelNotFound
}

func NewLocalConfigProvider(options config.DiscoveryConfigOpts, tlsMapper tlsConfigMapper) (api.DiscoveryProvider, error) {
	var channels []config.DiscoveryChannel
	if err := mapstructure.Decode(options, &channels); err != nil {
		return nil, errors.Wrap(err, `failed to decode params`)
	}

	return &LocalConfigProvider{channels: channels, tlsMapper: tlsMapper}, nil
}
