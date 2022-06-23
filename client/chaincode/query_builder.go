package chaincode

import (
	"context"
	"encoding/json"
	"fmt"

	fabricPeer "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"

	"github.com/atomyze-ru/hlf-sdk-go/api"
	"github.com/atomyze-ru/hlf-sdk-go/client/tx"
)

type QueryBuilder struct {
	channel       string
	chaincode     string
	fn            string
	args          [][]byte
	identity      msp.SigningIdentity
	peerPool      api.PeerPool
	transientArgs api.TransArgs
}

func (q *QueryBuilder) WithIdentity(identity msp.SigningIdentity) api.ChaincodeQueryBuilder {
	q.identity = identity
	return q
}

func (q *QueryBuilder) WithArguments(argBytes [][]byte) api.ChaincodeQueryBuilder {
	q.args = argBytes
	return q
}

// AsBytes TODO: think about interface in one style with Invoke
func (q *QueryBuilder) AsBytes(ctx context.Context) ([]byte, error) {
	if response, err := q.AsProposalResponse(ctx); err != nil {
		return nil, fmt.Errorf(`get proposal response: %w`, err)
	} else {
		return response.Response.Payload, nil
	}
}

func (q *QueryBuilder) AsJSON(ctx context.Context, out interface{}) error {
	if bytes, err := q.AsBytes(ctx); err != nil {
		return err
	} else {
		if err = json.Unmarshal(bytes, out); err != nil {
			return fmt.Errorf(`unmarshal JSON: %w`, err)
		}
	}
	return nil
}

func (q *QueryBuilder) AsProposalResponse(ctx context.Context) (*fabricPeer.ProposalResponse, error) {
	proposal, _, err := tx.Endorsement{
		Channel:      q.channel,
		Chaincode:    q.chaincode,
		Args:         tx.FnArgs(q.fn, q.args...),
		Signer:       q.identity,
		TransientMap: q.transientArgs,
	}.SignedProposal()

	if err != nil {
		return nil, fmt.Errorf(`create peer proposal: %w`, err)
	}

	return q.peerPool.EndorseOnMSP(ctx, q.identity.GetMSPIdentifier(), proposal)
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
		channel:   ccCore.channelName,
		chaincode: ccCore.name,
		fn:        fn,
		args:      tx.StringArgsBytes(args...),
		identity:  identity,
		peerPool:  ccCore.peerPool,
	}

	return q
}
