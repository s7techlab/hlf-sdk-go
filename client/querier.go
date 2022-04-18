package client

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"

	"github.com/s7techlab/hlf-sdk-go/api"
)

type QuerierDecorator struct {
	Querier       api.Querier
	DefaultSigner msp.SigningIdentity
}

func (q QuerierDecorator) Query(ctx context.Context, channel string, chaincode string, args [][]byte, signer msp.SigningIdentity, transient map[string][]byte) (*peer.Response, error) {

	if signer == nil {
		signer = q.DefaultSigner
	}

	return q.Querier.Query(ctx, channel, chaincode, args, signer, transient)
}
