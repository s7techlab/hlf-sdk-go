package util

import (
	"crypto/x509"

	"github.com/s7techlab/hlf-sdk-go/identity"
)

// Deprecated: use identity.LoadKeyPairFromMSP
// legacy method. loads ONLY cert from signcerts dir
func LoadKeyPairFromMSP(mspPath string) (*x509.Certificate, interface{}, error) {
	return identity.LoadKeyPairFromMSP(mspPath)
}

// Deprecated: use identity.LoadKeypairByCert
func LoadKeypairByCert(mspPath string, certRawBytes []byte) (*x509.Certificate, interface{}, error) {
	return identity.LoadKeypairByCert(mspPath, certRawBytes)
}
