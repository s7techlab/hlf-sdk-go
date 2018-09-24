package chaincode

import (
	"context"
	"encoding/json"

	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/peer"
)

type QueryBuilder struct {
	ccCore    *Core
	fn        string
	args      []string
	identity  msp.SigningIdentity
	processor api.PeerProcessor
	peerPool  api.PeerPool
}

func (q *QueryBuilder) WithIdentity(identity msp.SigningIdentity) api.ChaincodeQueryBuilder {
	q.identity = identity
	return q
}

// TODO: think about interface in one style with Invoke
func (q *QueryBuilder) AsBytes(ctx context.Context) ([]byte, error) {
	ccDef, err := q.ccCore.dp.Chaincode(q.ccCore.channelName, q.ccCore.name)
	if err != nil {
		return nil, errors.Wrap(err, `failed to get chaincode definition from discovery provider`)
	}

	proposal, _, err := q.processor.CreateProposal(ccDef, q.identity, q.fn, argsToBytes(q.args...))
	if err != nil {
		return nil, errors.Wrap(err, `failed to create peer proposal`)
	}

	// query only on local peer
	if respList, err := q.processor.Send(ctx, proposal, ccDef, q.peerPool); err != nil {
		return nil, errors.Wrap(err, `failed to get proposal response from peers`)
	} else {
		return respList[0].Response.Payload, nil
	}

	return nil, nil
}

func (q *QueryBuilder) AsJSON(ctx context.Context, out interface{}) error {
	if bytes, err := q.AsBytes(ctx); err != nil {
		return err
	} else {
		if err = json.Unmarshal(bytes, out); err != nil {
			return errors.Wrap(err, `failed to unmarshal JSON`)
		}
	}
	return nil
}

func NewQueryBuilder(ccCore *Core, identity msp.SigningIdentity, fn string, args ...string) *QueryBuilder {
	peerProcessor := peer.NewProcessor(ccCore.channelName)
	return &QueryBuilder{ccCore: ccCore, fn: fn, args: args, identity: identity, processor: peerProcessor, peerPool: ccCore.peerPool}
}
