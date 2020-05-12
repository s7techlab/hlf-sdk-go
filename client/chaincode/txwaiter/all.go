package txwaiter

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/util"
)

func All(cfg *api.DoOptions) (api.TxWaiter, error) {
	mspIds, err := util.GetMSPFromPolicy(cfg.DiscoveryChaincode.Policy)
	if err != nil {
		return nil, err
	}

	waiter := &allMspWaiter{
		onceSet: new(sync.Once),
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

	return waiter, nil
}

type allMspWaiter struct {
	delivers []api.DeliverClient
	onceSet  *sync.Once
	hasErr   bool
}

func (w *allMspWaiter) setErr() {
	w.onceSet.Do(func() { w.hasErr = true })
}

func (w *allMspWaiter) Wait(ctx context.Context, channel string, txid api.ChaincodeTx) error {
	var (
		wg   = new(sync.WaitGroup)
		errS = make(chan error, len(w.delivers))
	)

	for i := range w.delivers {
		wg.Add(1)
		go func(j int) {
			err := waitPerOne(ctx, w.delivers[j], channel, txid)
			if err != nil {
				w.setErr()
				errS <- err
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	close(errS)

	if w.hasErr {
		mErr := &api.MultiError{}
		for e := range errS {
			if e != nil {
				mErr.Errors = append(mErr.Errors, e)
			}
		}
		return mErr
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
