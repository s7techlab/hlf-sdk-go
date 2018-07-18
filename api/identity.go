package api

import (
	"github.com/hyperledger/fabric/msp"
)

type Identity interface {
	// GetSigningIdentity returns signing identity which will use presented crypto suite
	GetSigningIdentity(cs CryptoSuite) msp.SigningIdentity
}
