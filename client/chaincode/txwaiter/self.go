package txwaiter

import (
	"context"

	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/api"
)

// Self - default tx waiter and used on invoke flow
// txwaiter.Self  make subscribe tx on one peer endorser organization
func Self(cfg *api.DoOptions) (api.TxWaiter, error) {
	return &selfPeerWaiter{
		pool:     cfg.Pool,
		identity: cfg.Identity,
	}, nil
}

type selfPeerWaiter struct {
	pool     api.PeerPool
	identity msp.SigningIdentity
}

// Wait - implementation of api.TxWaiter interface
func (w *selfPeerWaiter) Wait(ctx context.Context, channel string, txid api.ChaincodeTx) error {
	mspID := w.identity.GetMSPIdentifier()
	deliver, err := w.pool.DeliverClient(mspID, w.identity)
	if err != nil {
		return errors.Wrapf(err, "%s: failed to get delivery client", mspID)
	}
	sub, err := deliver.SubscribeTx(ctx, channel, txid)
	if err != nil {
		return errors.Wrapf(err, "%s: failed to subscribe on tx event", mspID)
	}
	defer sub.Close()

	_, err = sub.Result()
	return err
}
