package identity

import (
	"crypto/x509"

	"github.com/s7techlab/hlf-sdk-go/api"
)

// Deprecated: use FromCertKeyPath
func NewMSPIdentity(mspId string, certPath string, keyPath string) (api.Identity, error) {
	return FromCertKeyPath(mspId, certPath, keyPath)
}

// Deprecated: use FromBytes
func NewMSPIdentityBytes(mspId string, certBytes []byte, keyBytes []byte) (api.Identity, error) {
	return FromBytes(mspId, certBytes, keyBytes)
}

// Deprecated: use FromMSPPath
func NewMSPIdentityFromPath(mspId string, mspPath string) (api.Identity, error) {
	return FromMSPPath(mspId, mspPath)
}

// Deprecated: use New
func NewMSPIdentityRaw(mspId string, cert *x509.Certificate, privateKey interface{}) (api.Identity, error) {
	return New(mspId, cert, privateKey)
}

// Deprecated: use CollectionFromMSPPath
func NewMSPIdentitiesFromPath(mspID string, mspPath string) (*MSP, error) {
	return CollectionFromMSPPath(mspID, mspPath)
}

// Deprecated: use New
func NewEnrollIdentity(privateKey interface{}) (api.Identity, error) {
	return &identity{privateKey: privateKey}, nil
}
