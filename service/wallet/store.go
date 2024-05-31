package wallet

import (
	walletproto "github.com/s7techlab/hlf-sdk-go/proto/wallet"
)

type (
	Store interface {
		Get(label string) (*walletproto.IdentityInWallet, error)
		Set(identity *walletproto.IdentityInWallet) error
		List() (labels []string, err error)
		Delete(label string) error
	}
)
