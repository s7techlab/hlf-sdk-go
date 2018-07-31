package api

import "github.com/s7techlab/hlf-sdk-go/api/config"

// CryptoSuite describes common cryptographic operations
type CryptoSuite interface {
	// Sign is used for signing message by presented private key
	Sign(msg []byte, key interface{}) ([]byte, error)
	// Verify is used for verifying signature for presented message and public key
	Verify(publicKey interface{}, msg, sig []byte) error
	// Hash is used for hashing presented data
	Hash(data []byte) []byte
	// NewPrivateKey generates new private key
	NewPrivateKey() (interface{}, error)
	// Initialize is used for suite instantiation using presented options
	Initialize(opts config.CryptoSuiteOpts) (CryptoSuite, error)
}
