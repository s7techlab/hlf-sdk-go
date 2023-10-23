package txwaiter

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/api"
	clienterr "github.com/s7techlab/hlf-sdk-go/client/errors"
)

// All - need use on invoke flow for check transaction codes for each organization from endorsement policy
// txwaiter.All  will be made to subscribe tx for each of the peer organizations from the endorsement policy
func All(cfg *api.DoOptions) (api.TxWaiter, error) {
	waiter := &allMspWaiter{
		onceSet: new(sync.Once),
	}

	// make delivers for each mspID
	errD := new(clienterr.MultiError)
	for i := range cfg.EndorsingMspIDs {
		peerDeliver, err := cfg.Pool.DeliverClient(cfg.EndorsingMspIDs[i], cfg.Identity)
		if err != nil {
			errD.Add(errors.Wrapf(err, "%s: failed to get delivery client", cfg.EndorsingMspIDs[i]))
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

// Wait - implementation of api.TxWaiter interface
func (w *allMspWaiter) Wait(ctx context.Context, channel string, txId string) error {
	var (
		wg   = new(sync.WaitGroup)
		errS = make(chan error, len(w.delivers))
	)

	for i := range w.delivers {
		wg.Add(1)
		go func(j int) {
			err := waitPerOne(ctx, w.delivers[j], channel, txId)
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
		mErr := &clienterr.MultiError{}
		for e := range errS {
			if e != nil {
				mErr.Errors = append(mErr.Errors, e)
			}
		}
		return mErr
	}

	return nil
}

func waitPerOne(ctx context.Context, deliver api.DeliverClient, channelName string, txId string) error {
	sub, err := deliver.SubscribeTx(ctx, channelName, txId)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe on tx event")
	}
	defer func() { _ = sub.Close() }()

	_, err = sub.Result()
	return err
}
