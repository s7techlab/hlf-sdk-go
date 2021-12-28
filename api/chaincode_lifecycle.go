package api

import (
	"context"

	lb "github.com/hyperledger/fabric-protos-go/peer/lifecycle"
)

type Lifecycle interface {
	QueryInstalledChaincodes(ctx context.Context) (*lb.QueryInstalledChaincodesResult, error)
}
