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
	return SignerFromMSPPath(mspId, mspPath)
}

// Deprecated: use New
func NewMSPIdentityRaw(mspId string, cert *x509.Certificate, privateKey interface{}) (api.Identity, error) {
	return New(mspId, cert, privateKey), nil
}

// Deprecated: use MSPFromPath to create MSP
func NewMSPIdentitiesFromPath(mspID string, mspPath string) (*MSPConfig, error) {
	return MSPFromPath(mspID, mspPath)
}

// Deprecated: use New
func NewEnrollIdentity(privateKey interface{}) (api.Identity, error) {
	return &identity{privateKey: privateKey}, nil
}

// Deprecated: use SignerFromMSPPath to create identity
// LoadKeyPairFromMSP - legacy method. loads ONLY cert from signcerts dir
func LoadKeyPairFromMSP(mspPath string) (*x509.Certificate, interface{}, error) {
	identity, err := SignerFromMSPPath(``, mspPath)
	if err != nil {
		return nil, nil, err
	}

	return identity.certificate, identity.privateKey, nil
}
