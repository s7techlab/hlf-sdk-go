package api

import (
	"context"

	lifecycle3 "github.com/hyperledger/fabric-protos-go/peer/lifecycle"
)

type Lifecycle interface {
	QueryInstalledChaincodes(ctx context.Context) (*lifecycle3.QueryInstalledChaincodesResult, error)
}
