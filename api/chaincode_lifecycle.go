// +build !fabric2

package api

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/peer"
)

type Lifecycle interface {
	QueryInstalledChaincodes(ctx context.Context) (*peer.ChaincodeQueryResponse, error)
}
