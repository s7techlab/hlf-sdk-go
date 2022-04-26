package tx

import (
	"context"
	"errors"

	"github.com/hyperledger/fabric/msp"
)

var (
	// ErrSignerNotDefinedInContext msp.SigningIdentity is not defined in context
	ErrSignerNotDefinedInContext = errors.New(`signer is not defined in context`)
)

const (
	CtxTransientKey    = `TransientMap`
	CtxSignerKey       = `SigningIdentity`
	CtxTxWaiterKey     = `TxWaiter`
	CtxEndorserMSPsKey = `EndorserMSPs`
)

func ContextWithTransientMap(ctx context.Context, transient map[string][]byte) context.Context {
	return context.WithValue(ctx, CtxTransientKey, transient)
}

func ContextWithTransientValue(ctx context.Context, key string, value []byte) context.Context {
	transient, ok := ctx.Value(CtxTransientKey).(map[string][]byte)
	if !ok {
		transient = make(map[string][]byte)
	}
	transient[key] = value
	return context.WithValue(ctx, CtxTransientKey, transient)
}

func TransientFromContext(ctx context.Context) map[string][]byte {
	if transient, ok := ctx.Value(CtxTransientKey).(map[string][]byte); ok {
		return transient
	}

	return nil
}

func ContextWithSigner(ctx context.Context, signer msp.SigningIdentity) context.Context {
	return context.WithValue(ctx, CtxSignerKey, signer)
}

func SignerFromContext(ctx context.Context) msp.SigningIdentity {
	if signer, ok := ctx.Value(CtxSignerKey).(msp.SigningIdentity); ok {
		return signer
	}

	return nil
}

func ContextWithTxWaiter(ctx context.Context, txWaiterType string) context.Context {
	return context.WithValue(ctx, CtxTxWaiterKey, txWaiterType)
}

// TxWaiterFromContext - fetch 'txWaiterType' param which identify transaction waiting policy
// what params you'll have depends on your implementation
// for example, in hlf-sdk:
// available: 'self'(wait for one peer of endorser org), 'all'(wait for each organizations from endorsement policy)
// default is 'self'(even if you pass empty string)
func TxWaiterFromContext(ctx context.Context) string {
	txWaiter, _ := ctx.Value(CtxTxWaiterKey).(string)
	return txWaiter
}

func ContextWithEndorserMSPs(ctx context.Context, endorserMSPs []string) context.Context {
	return context.WithValue(ctx, CtxEndorserMSPsKey, endorserMSPs)
}

func EndorserMSPsFromContext(ctx context.Context) []string {
	if endorserMSPs, ok := ctx.Value(CtxEndorserMSPsKey).([]string); ok {
		return endorserMSPs
	}
	return nil
}
