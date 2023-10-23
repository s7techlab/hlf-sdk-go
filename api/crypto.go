package api

import (
	"github.com/s7techlab/hlf-sdk-go/crypto"
	"github.com/s7techlab/hlf-sdk-go/crypto/ecdsa"
)

func DefaultCryptoSuite() crypto.CryptoSuite {
	suite, _ := crypto.GetSuite(ecdsa.Module, ecdsa.DefaultOpts)
	return suite
}
