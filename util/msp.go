package util

import (
	"crypto/x509"

	"github.com/atomyze-ru/hlf-sdk-go/identity"
)

// Deprecated: use identity.SignerFromMSPPath to create identity
// legacy method. loads ONLY cert from signcerts dir
func LoadKeyPairFromMSP(mspPath string) (*x509.Certificate, interface{}, error) {
	return identity.LoadKeyPairFromMSP(mspPath)
}

// Deprecated: use identity.KeyPairForCert
func LoadKeypairByCert(mspPath string, certRaw []byte) (*x509.Certificate, interface{}, error) {
	return identity.KeyPairForCert(certRaw, identity.KeystorePath(mspPath))
}
