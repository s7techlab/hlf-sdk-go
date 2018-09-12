package chaincode

import (
	"context"
	"github.com/hyperledger/fabric/msp"
	"github.com/s7techlab/hlf-sdk-go/api"
)

type Core struct {
	name          string
	channelName   string
	peer          api.Peer
	orderer       api.Orderer
	dp            api.DiscoveryProvider
	identity      msp.SigningIdentity
	deliverClient api.DeliverClient
}

func (c *Core) Invoke(fn string) api.ChaincodeInvokeBuilder {
	return NewInvokeBuilder(c, fn)
}

func (c *Core) Query(fn string, args ...string) api.ChaincodeQueryBuilder {
	return NewQueryBuilder(c, c.identity, fn, args...)
}

func (c *Core) Install(version string) {

	panic("implement me")
}

func (c *Core) Subscribe(ctx context.Context, seekOption ...api.EventCCSeekOption) api.EventCCSubscription {
	return c.deliverClient.SubscribeCC(ctx, c.channelName, c.name, seekOption...)
}

func NewCore(
	ccName, channelName string,
	peer api.Peer,
	orderer api.Orderer,
	dp api.DiscoveryProvider,
	identity msp.SigningIdentity,
	deliverClient api.DeliverClient,
) *Core {
	return &Core{
		name:          ccName,
		channelName:   channelName,
		peer:          peer,
		orderer:       orderer,
		dp:            dp,
		identity:      identity,
		deliverClient: deliverClient,
	}
}
