package api

import (
	"crypto/x509"

	"github.com/hyperledger/fabric/msp"

	"github.com/s7techlab/hlf-sdk-go/crypto"
)

type Identity interface {
	// GetSigningIdentity returns signing identity which will use presented crypto suite
	GetSigningIdentity(cs crypto.CryptoSuite) msp.SigningIdentity
	// GetMSPIdentifier return msp id
	GetMSPIdentifier() string
	// GetPEM returns certificate in PEM format
	GetPEM() []byte
	// GetCert returns X509 Certificate
	GetCert() *x509.Certificate
}
