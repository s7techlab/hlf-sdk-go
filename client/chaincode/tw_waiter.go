package chaincode

import (
	"github.com/s7techlab/hlf-sdk-go/api"
)

type TxWaitBuilder func(cfg *api.DoOptions) (api.TxWaiter, error)

func WithTxWaiter(builder TxWaitBuilder) api.DoOption {
	return func(cfg *api.DoOptions) (err error) {
		cfg.TxWaiter, err = builder(cfg)
		return
	}
}
