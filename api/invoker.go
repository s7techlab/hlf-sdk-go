package api

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
)

// Invoker interface describes common operations for chaincode
type Invoker interface {
	// Invoke method allows invoking chaincode
	Invoke(ctx context.Context, from msp.SigningIdentity, channel string, chaincode string, fn string, args [][]byte, transArgs TransArgs, doOpts ...DoOption) (*peer.Response, ChaincodeTx, error)
	// Query method allows querying chaincode without sending response to orderer
	Query(ctx context.Context, from msp.SigningIdentity, channel string, chaincode string, fn string, args [][]byte, transArgs TransArgs) (*peer.Response, error)
	// Subscribe allows subscribing on chaincode events
	Subscribe(ctx context.Context, from msp.SigningIdentity, channel, chaincode string) (EventCCSubscription, error)
}
