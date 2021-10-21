package invoker

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/v2/api"
)

type invoker struct {
	core api.Core
}

func (i *invoker) Invoke(
	ctx context.Context,
	from msp.SigningIdentity,
	channel string,
	chaincode string,
	fn string,
	args [][]byte,
	transArgs api.TransArgs,
	doOpts ...api.DoOption,
) (*peer.Response, api.ChaincodeTx, error) {
	ссConnection, err := i.core.
		Channel(channel).
		Chaincode(ctx, chaincode)
	if err != nil {
		return nil, "", err
	}

	return ссConnection.Invoke(fn).
		WithIdentity(from).
		ArgBytes(args).
		Transient(transArgs).
		Do(ctx, doOpts...)
}

func (i *invoker) Query(ctx context.Context, from msp.SigningIdentity, channel string, chaincode string, fn string, args [][]byte, transArgs api.TransArgs) (*peer.Response, error) {
	argSs := make([]string, 0)
	for _, arg := range args {
		argSs = append(argSs, string(arg))
	}

	ссConnection, err := i.core.
		Channel(channel).
		Chaincode(ctx, chaincode)
	if err != nil {
		return nil, err
	}

	if resp, err := ссConnection.Query(fn, argSs...).WithIdentity(from).Transient(transArgs).AsProposalResponse(ctx); err != nil {
		return nil, errors.Wrap(err, `failed to query chaincode`)
	} else {
		return resp.Response, nil
	}
}

func (i *invoker) Subscribe(ctx context.Context, from msp.SigningIdentity, channel, chaincode string) (api.EventCCSubscription, error) {
	ссConnection, err := i.core.
		Channel(channel).
		Chaincode(ctx, chaincode)
	if err != nil {
		return nil, err
	}

	return ссConnection.Subscribe(ctx)
}

func New(core api.Core) api.Invoker {
	return &invoker{core: core}
}
