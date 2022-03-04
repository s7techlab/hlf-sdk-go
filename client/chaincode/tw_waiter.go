package chaincode

import (
	"github.com/s7techlab/hlf-sdk-go/api"
)

// TxWaitBuilder function signature for pluggable setter on Do options
type TxWaitBuilder func(cfg *api.DoOptions) (api.TxWaiter, error)

// WithTxWaiter - add option for set custom tx waiter
func WithTxWaiter(builder TxWaitBuilder) api.DoOption {
	return func(cfg *api.DoOptions) (err error) {
		cfg.TxWaiter, err = builder(cfg)
		return
	}
}
