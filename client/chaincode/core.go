package chaincode

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric/msp"

	"github.com/s7techlab/hlf-sdk-go/api"
)

type Core struct {
	mspId         string
	name          string
	channelName   string
	endorsingMSPs []string
	peerPool      api.PeerPool
	orderer       api.Orderer

	identity msp.SigningIdentity
}

func NewCore(
	mspId,
	ccName,
	channelName string,
	endorsingMSPs []string,
	peerPool api.PeerPool,
	orderer api.Orderer,
	identity msp.SigningIdentity,
) *Core {
	return &Core{
		mspId:         mspId,
		name:          ccName,
		channelName:   channelName,
		endorsingMSPs: endorsingMSPs,
		peerPool:      peerPool,
		orderer:       orderer,
		identity:      identity,
	}
}

func (c *Core) GetPeers() []api.Peer {
	peers := make([]api.Peer, 0)

	peersMap := c.peerPool.GetPeers()
	for _, endorsingMSP := range c.endorsingMSPs {
		if ps, ok := peersMap[endorsingMSP]; ok {
			peers = append(peers, ps...)
		}
	}

	return peers
}

func (c *Core) Invoke(fn string) api.ChaincodeInvokeBuilder {
	return NewInvokeBuilder(c, fn)
}

func (c *Core) Query(fn string, args ...string) api.ChaincodeQueryBuilder {
	return NewQueryBuilder(c, c.identity, fn, args...)
}

func (c *Core) Subscribe(ctx context.Context) (api.EventCCSubscription, error) {
	peerDeliver, err := c.peerPool.DeliverClient(c.mspId, c.identity)
	if err != nil {
		return nil, fmt.Errorf(`initiate DeliverClient: %w`, err)
	}
	return peerDeliver.SubscribeCC(ctx, c.channelName, c.name)
}
