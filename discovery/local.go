package discovery

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric/common/policydsl"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/atomyze-ru/hlf-sdk-go/api"
	"github.com/atomyze-ru/hlf-sdk-go/api/config"
)

// implementation of api.DiscoveryProvider interface
var _ api.DiscoveryProvider = (*LocalConfigProvider)(nil)

type LocalConfigProvider struct {
	tlsMapper tlsConfigMapper
	channels  []config.DiscoveryChannel `yaml:"channels"`
}
type opts struct {
	Channels []config.DiscoveryChannel `yaml:"channels"`
}

func NewLocalConfigProvider(options config.DiscoveryConfigOpts, tlsMapper tlsConfigMapper) (api.DiscoveryProvider, error) {
	var opts opts
	if err := mapstructure.Decode(options, &opts); err != nil {
		return nil, errors.Wrap(err, `failed to decode params`)
	}

	return &LocalConfigProvider{channels: opts.Channels, tlsMapper: tlsMapper}, nil
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
					msps, err := getMSPsFromPolicy(cc.Policy)
					if err != nil {
						return nil, err
					}

					for i := range msps {
						mspID := msps[i]
						hostAddr := "" // no addr in channel config, peer must be already in pool
						ccDTO.addEndpointToEndorsers(mspID, hostAddr)
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

func (d *LocalConfigProvider) LocalPeers(_ context.Context) (api.LocalPeersDiscoverer, error) {
	return nil, fmt.Errorf("LocalPeers for LocalConfigProvider not implemented")
}

func getMSPsFromPolicy(policy string) ([]string, error) {
	policyEnvelope, err := policydsl.FromString(policy)
	if err != nil {
		return nil, errors.Wrap(err, `failed to parse policy`)
	}

	mspIds := make([]string, 0)

	for _, id := range policyEnvelope.Identities {
		var mspIdentity msp.SerializedIdentity
		if err = proto.Unmarshal(id.Principal, &mspIdentity); err != nil {
			return nil, errors.Wrap(err, `failed to get MSP identity`)
		} else {
			mspIds = append(mspIds, mspIdentity.Mspid)
		}
	}

	return mspIds, nil
}
