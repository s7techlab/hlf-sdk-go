package chaincode

import (
	"context"

	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/v2/api"
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

func (c *Core) Subscribe(ctx context.Context) (api.EventCCSubscription, error) {
	peerDeliver, err := c.peerPool.DeliverClient(c.mspId, c.identity)
	if err != nil {
		return nil, errors.Wrap(err, `failed to initiate DeliverClient`)
	}
	return peerDeliver.SubscribeCC(ctx, c.channelName, c.name)
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
