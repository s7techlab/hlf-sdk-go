package chaincode

import (
	"context"

	"github.com/hyperledger/fabric/msp"
	"github.com/s7techlab/hlf-sdk-go/api"
)

type Core struct {
	mspId       string
	name        string
	channelName string
	peerPool    api.PeerPool
	orderer     api.Orderer
	dp          api.DiscoveryProvider
	identity    msp.SigningIdentity
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
	peerDeliver, _ := c.peerPool.DeliverClient(c.mspId, c.identity)
	return peerDeliver.SubscribeCC(ctx, c.channelName, c.name, seekOption...)
}

func NewCore(mspId, ccName, channelName string, peerPool api.PeerPool, orderer api.Orderer, dp api.DiscoveryProvider, identity msp.SigningIdentity) *Core {
	return &Core{
		mspId:       mspId,
		name:        ccName,
		channelName: channelName,
		peerPool:    peerPool,
		orderer:     orderer,
		dp:          dp,
		identity:    identity,
	}
}
