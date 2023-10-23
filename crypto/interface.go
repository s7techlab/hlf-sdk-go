package crypto

import (
	"crypto/x509"
)

// Suite describes common cryptographic operations
type Suite interface {
	// Sign is used for signing message by presented private key
	Sign(msg []byte, key interface{}) ([]byte, error)
	// Verify is used for verifying signature for presented message and public key
	Verify(publicKey interface{}, msg, sig []byte) error
	// Hash is used for hashing presented data
	Hash(data []byte) []byte
	// NewPrivateKey generates new private key
	NewPrivateKey() (interface{}, error)
	// GetSignatureAlgorithm returns signature algorithm
	GetSignatureAlgorithm() x509.SignatureAlgorithm
}
