package api

import (
	"context"
)

type Channel interface {
	// Chaincode returns chaincode instance by chaincode name
	Chaincode(ctx context.Context, name string) (Chaincode, error)
	// Join channel
	Join(ctx context.Context) error
}

// types which identify tx "wait'er" policy
// we don't make it as alias for preventing binding to our lib
const (
	TxWaiterSelfType string = "self"
	TxWaiterAllType  string = "all"
)
