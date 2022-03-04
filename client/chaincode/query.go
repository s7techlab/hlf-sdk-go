package chaincode

import (
	"context"
	"encoding/json"

	fabricPeer "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/peer"
)

type QueryBuilder struct {
	cc            string
	fn            string
	argBytes      [][]byte
	identity      msp.SigningIdentity
	processor     api.PeerProcessor
	peerPool      api.PeerPool
	transientArgs api.TransArgs
}

func (q *QueryBuilder) WithIdentity(identity msp.SigningIdentity) api.ChaincodeQueryBuilder {
	q.identity = identity

	return q
}

func (q *QueryBuilder) WithArguments(argBytes [][]byte) api.ChaincodeQueryBuilder {
	q.argBytes = argBytes

	return q
}

// AsBytes TODO: think about interface in one style with Invoke
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
	proposal, _, err := q.processor.CreateProposal(q.cc, q.identity, q.fn, q.argBytes, q.transientArgs)
	if err != nil {
		return nil, errors.Wrap(err, `failed to create peer proposal`)
	}

	return q.peerPool.Process(ctx, q.identity.GetMSPIdentifier(), proposal)
}

// Do makes invoke with built arguments
func (q *QueryBuilder) Do(ctx context.Context) (*fabricPeer.Response, error) {
	res, err := q.AsProposalResponse(ctx)
	if err != nil {
		return nil, err
	}

	return res.Response, nil
}

func (q *QueryBuilder) Transient(args api.TransArgs) api.ChaincodeQueryBuilder {
	q.transientArgs = args

	return q
}

func NewQueryBuilder(ccCore *Core, identity msp.SigningIdentity, fn string, args ...string) api.ChaincodeQueryBuilder {
	q := &QueryBuilder{
		cc:        ccCore.name,
		fn:        fn,
		argBytes:  argsToBytes(args...),
		identity:  identity,
		processor: peer.NewProcessor(ccCore.channelName),
		peerPool:  ccCore.peerPool,
	}

	return q
}
