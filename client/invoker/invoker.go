package invoker

import (
	"context"

	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
)

type invoker struct {
	core api.Core
}

const waitOfAllMspContextKey = `WaitOfAllMsp`

func WaitOfAllMsp(ctx context.Context) context.Context {
	return context.WithValue(ctx, waitOfAllMspContextKey, true)
}

func hasWaitOfAllMsp(ctx context.Context) bool {
	return ctx.Value(waitOfAllMspContextKey) != nil
}

func (i *invoker) Invoke(ctx context.Context, from msp.SigningIdentity, channel string, chaincode string, fn string, args [][]byte, transArgs api.TransArgs) (*peer.Response, api.ChaincodeTx, error) {
	return i.core.Channel(channel).Chaincode(chaincode).Invoke(fn).WithIdentity(from).WithWaitTxForAllMsp(hasWaitOfAllMsp(ctx)).ArgBytes(args).Transient(transArgs).Do(ctx)
}

func (i *invoker) Query(ctx context.Context, from msp.SigningIdentity, channel string, chaincode string, fn string, args [][]byte, transArgs api.TransArgs) (*peer.Response, error) {
	argSs := make([]string, 0)
	for _, arg := range args {
		argSs = append(argSs, string(arg))
	}

	if resp, err := i.core.Channel(channel).Chaincode(chaincode).Query(fn, argSs...).WithIdentity(from).Transient(transArgs).AsProposalResponse(ctx); err != nil {
		return nil, errors.Wrap(err, `failed to query chaincode`)
	} else {
		return resp.Response, nil
	}
}

func (i *invoker) Subscribe(ctx context.Context, from msp.SigningIdentity, channel, chaincode string) (api.EventCCSubscription, error) {
	return i.core.Channel(channel).Chaincode(chaincode).Subscribe(ctx)
}

func New(core api.Core) api.Invoker {
	return &invoker{core: core}
}
