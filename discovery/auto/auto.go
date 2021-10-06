package auto

import (
	"context"
	"log"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/discovery"
)

const Name = `auto`

type DiscoveryProvider struct {
	core api.Core
}

func (d *DiscoveryProvider) Channels() ([]api.DiscoveryChannel, error) {
	// if len(d.opts.Channels) > 0 {
	// 	return d.opts.Channels, nil
	// }
	// return nil, discovery.ErrNoChannels
	return nil, nil
}

func (d *DiscoveryProvider) Channel(channelName string) (*api.DiscoveryChannel, error) {
	chanConfig, err := d.core.System().CSCC().GetChannelConfig(context.Background(), channelName)
	if err != nil {
		return nil, err
	}
	log.Printf("svdf %+v", chanConfig)

	ordererAdderesses, err := getOrdererAddresses(chanConfig)
	if err != nil {
		return nil, err
	}

	orderersRes := make([]config.ConnectionConfig, len(ordererAdderesses))
	for i := range ordererAdderesses {
		orderersRes[i].Host = ordererAdderesses[i]
	}

	res := &api.DiscoveryChannel{
		Name:     channelName,
		Orderers: orderersRes,
	}

	return res, nil
}

func (d *DiscoveryProvider) Chaincode(channelName string, ccName string) (*api.DiscoveryChaincode, error) {
	// var channelFoundFlag bool

	// for _, ch := range d.opts.Channels {
	// 	if ch.Name == channelName {
	// 		channelFoundFlag = true
	// 		for _, cc := range ch.Chaincodes {
	// 			if cc.Name == ccName {
	// 				return &cc, nil
	// 			}
	// 		}
	// 	}
	// }

	// if channelFoundFlag {
	// 	return nil, discovery.ErrNoChaincodes
	// }
	// return nil, discovery.ErrChannelNotFound
	return &api.DiscoveryChaincode{Name: ccName, Type: "golang"}, nil
}

func (d *DiscoveryProvider) Chaincodes(channelName string) ([]api.DiscoveryChaincode, error) {
	// for _, ch := range d.opts.Channels {
	// 	if ch.Name == channelName {
	// 		return ch.Chaincodes, nil
	// 	}
	// }
	// return nil, discovery.ErrChannelNotFound
	return nil, nil
}

func (d *DiscoveryProvider) Initialize(c config.DiscoveryConfigOpts, a api.PeerPool, core api.Core) (api.DiscoveryProvider, error) {
	d.core = core
	return d, nil
}

func init() {
	discovery.Register(Name, &DiscoveryProvider{})
}

/*
chansResponse, err := d.core.System().CSCC().GetChannels(context.Background())
	if err != nil {
		return nil, err
	}

	isProvidedChannelConnected := func(chans []*peer.ChannelInfo, cn string) bool {
		for i := range chansResponse.Channels {
			chanID := chansResponse.Channels[i].ChannelId
			if channelName == chanID {
				return true
			}
		}
		return false
	}
	if !isProvidedChannelConnected(chansResponse.Channels, channelName) {
		return nil, discovery.ErrChannelNotFound
	}

	res := &api.DiscoveryChannel{
		Name: channelName,
		// TODO other data?
	}
*/
