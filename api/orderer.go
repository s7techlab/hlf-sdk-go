package api

import (
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/orderer"
)

type Orderer interface {
	// Broadcast sends envelope to orderer and returns it's result
	Broadcast(envelope *common.Envelope) (*orderer.BroadcastResponse, error)
	// Deliver fetches block from orderer by envelope
	Deliver(envelope *common.Envelope) (*common.Block, error)
}
