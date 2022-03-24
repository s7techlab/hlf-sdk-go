package client

import (
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/crypto"
	"github.com/s7techlab/hlf-sdk-go/crypto/ecdsa"
)

func DefaultCryptoSuite() api.CryptoSuite {
	suite, _ := crypto.GetSuite(ecdsa.Module, ecdsa.DefaultOpts)
	return suite
}
