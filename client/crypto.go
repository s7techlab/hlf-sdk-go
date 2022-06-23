package client

import (
	"github.com/atomyze-ru/hlf-sdk-go/api"
	"github.com/atomyze-ru/hlf-sdk-go/crypto"
	"github.com/atomyze-ru/hlf-sdk-go/crypto/ecdsa"
)

func DefaultCryptoSuite() api.CryptoSuite {
	suite, _ := crypto.GetSuite(ecdsa.Module, ecdsa.DefaultOpts)
	return suite
}
