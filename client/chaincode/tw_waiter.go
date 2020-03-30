package chaincode

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/util"
)

type TxWaiterMode int

const (
	TxWaiterSelf TxWaiterMode = iota
	TxWaiterAll
)

func WithTxWaiter(mode TxWaiterMode) api.DoOption {
	return func(cfg *api.DoOptions) (err error) {
		cfg.TxWaiter, err = NewTxWaiter(cfg, mode)
		return
	}
}

func NewTxWaiter(cfg *api.DoOptions, mode TxWaiterMode) (api.TxWaiter, error) {
	var (
		mspIds []string
		waiter *txWaiter
	)

	// check mode
	// and fill mspIDS
	switch mode {
	case TxWaiterAll:
		mspIds, _ = util.GetMSPFromPolicy(cfg.DiscoveryChaincode.Policy)
	case TxWaiterSelf:
		mspIds = []string{cfg.Identity.GetMSPIdentifier()}
	default:
		return nil, errors.Errorf("Unsupported tx waiter mode %s", mode)
	}

	// make delivers for each mspID
	errD := new(api.MultiError)
	for i := range mspIds {
		peerDeliver, err := cfg.Pool.DeliverClient(mspIds[i], cfg.Identity)
		if err != nil {
			errD.Add(errors.Wrapf(err, "%s: failed to get delivery client", mspIds[i]))
			continue
		}

		waiter.delivers = append(waiter.delivers, peerDeliver)
	}
	if len(errD.Errors) != 0 {
		return nil, errD
	}

	waiter.onceSet = new(sync.Once)
	waiter.channelName = cfg.Channel

	return waiter, nil
}

type txWaiter struct {
	delivers    []api.DeliverClient
	channelName string
	onceSet     *sync.Once
	hasErr      bool
}

func (w *txWaiter) setErr() {
	w.onceSet.Do(func() { w.hasErr = true })
}

func (w *txWaiter) Wait(ctx context.Context, txid api.ChaincodeTx) error {
	if len(w.delivers) == 1 {
		return waitPerOne(ctx, w.delivers[0], w.channelName, txid)
	}

	var (
		wg   = new(sync.WaitGroup)
		errS = make([]error, len(w.delivers))
	)

	for i := range w.delivers {
		wg.Add(1)
		go func(j int) {
			err := waitPerOne(ctx, w.delivers[j], w.channelName, txid)
			if err != nil {
				w.setErr()
				errS[j] = err
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	if w.hasErr {
		return &api.MultiError{Errors: errS}
	}

	return nil
}

func waitPerOne(ctx context.Context, deliver api.DeliverClient, channelName string, txid api.ChaincodeTx) error {
	sub, err := deliver.SubscribeTx(ctx, channelName, txid)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe on tx event")
	}
	defer sub.Close()

	_, err = sub.Result()
	return err
}
