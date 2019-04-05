package chaincode

import (
	"context"
	"encoding/json"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/peer"
	"github.com/hyperledger/fabric/msp"
	fabricPeer "github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

type QueryBuilder struct {
	ccCore        *Core
	fn            string
	args          []string
	identity      msp.SigningIdentity
	processor     api.PeerProcessor
	peerPool      api.PeerPool
	transientArgs api.TransArgs
}

func (q *QueryBuilder) WithIdentity(identity msp.SigningIdentity) api.ChaincodeQueryBuilder {
	q.identity = identity
	return q
}

// TODO: think about interface in one style with Invoke
func (q *QueryBuilder) AsBytes(ctx context.Context) ([]byte, error) {
	if response, err := q.AsProposalResponse(ctx); err != nil {
		return nil, errors.Wrap(err, `failed to get proposal response`)
	} else {
		return response.Response.Payload, nil
	}
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

func (q *QueryBuilder) AsProposalResponse(ctx context.Context) (*fabricPeer.ProposalResponse, error) {
	ccDef, err := q.ccCore.dp.Chaincode(q.ccCore.channelName, q.ccCore.name)
	if err != nil {
		return nil, errors.Wrap(err, `failed to get chaincode definition from discovery provider`)
	}

	proposal, _, err := q.processor.CreateProposal(ccDef, q.identity, q.fn, argsToBytes(q.args...), q.transientArgs)
	if err != nil {
		return nil, errors.Wrap(err, `failed to create peer proposal`)
	}

	return q.peerPool.Process(q.identity.GetMSPIdentifier(), ctx, proposal)
}

func (q *QueryBuilder) Transient(args api.TransArgs) api.ChaincodeQueryBuilder {
	q.transientArgs = args
	return q
}

func NewQueryBuilder(ccCore *Core, identity msp.SigningIdentity, fn string, args ...string) api.ChaincodeQueryBuilder {
	peerProcessor := peer.NewProcessor(ccCore.channelName)
	return &QueryBuilder{ccCore: ccCore, fn: fn, args: args, identity: identity, processor: peerProcessor, peerPool: ccCore.peerPool}
}
