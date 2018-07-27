package channel

import (
	"sync"

	"github.com/hyperledger/fabric/msp"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/member/chaincode"
)

type Core struct {
	name          string
	peer          api.Peer
	orderer       api.Orderer
	chaincodes    map[string]*chaincode.Core
	chaincodesMx  sync.Mutex
	dp            api.DiscoveryProvider
	identity      msp.SigningIdentity
	deliverClient api.DeliverClient
}

func (c *Core) Chaincode(name string) api.Chaincode {
	c.chaincodesMx.Lock()
	defer c.chaincodesMx.Unlock()
	if cc, ok := c.chaincodes[name]; !ok {
		cc = chaincode.NewCore(name, c.name, c.peer, c.orderer, c.dp, c.identity, c.deliverClient)
		c.chaincodes[name] = cc
		return cc
	} else {
		return cc
	}
}

func NewCore(name string, peer api.Peer, orderer api.Orderer, dp api.DiscoveryProvider, identity msp.SigningIdentity, deliverClient api.DeliverClient) api.Channel {
	return &Core{name: name, peer: peer, orderer: orderer, chaincodes: make(map[string]*chaincode.Core), dp: dp, identity: identity, deliverClient: deliverClient}
}
