package api

import (
	"context"

	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/peer"
)

// Invoker interface describes common operations for chaincode
type Invoker interface {
	// Invoke method allows to invoke chaincode
	Invoke(ctx context.Context, from msp.SigningIdentity, channel string, chaincode string, fn string, args [][]byte, transArgs TransArgs) (*peer.Response, ChaincodeTx, error)
	// Query method allows to query chaincode without sending response to orderer
	Query(ctx context.Context, from msp.SigningIdentity, channel string, chaincode string, fn string, args [][]byte, transArgs TransArgs) (*peer.Response, error)
	// Subscribe allows to subscribe on chaincode events
	Subscribe(ctx context.Context, from msp.SigningIdentity, channel, chaincode string) (EventCCSubscription, error)
}
