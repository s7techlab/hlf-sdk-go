package channel

import (
	"sync"

	"github.com/hyperledger/fabric/msp"
	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/client/chaincode"
)

type Core struct {
	mspId        string
	name         string
	peerPool     api.PeerPool
	orderer      api.Orderer
	chaincodes   map[string]*chaincode.Core
	chaincodesMx sync.Mutex
	dp           api.DiscoveryProvider
	identity     msp.SigningIdentity
	log          *zap.Logger
}

func (c *Core) Chaincode(name string) api.Chaincode {
	c.chaincodesMx.Lock()
	defer c.chaincodesMx.Unlock()
	if cc, ok := c.chaincodes[name]; !ok {
		cc = chaincode.NewCore(c.mspId, name, c.name, c.peerPool, c.orderer, c.dp, c.identity)
		c.chaincodes[name] = cc
		return cc
	} else {
		return cc
	}
}

func NewCore(mspId string, name string, peerPool api.PeerPool, orderer api.Orderer, dp api.DiscoveryProvider, identity msp.SigningIdentity, log *zap.Logger) api.Channel {
	return &Core{mspId: mspId, name: name, peerPool: peerPool, orderer: orderer, chaincodes: make(map[string]*chaincode.Core), dp: dp, identity: identity, log: log}
}
