package identity

/**
Package allows to use identities from wonderful gohfc package by CognitionFoundry
*/

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/CognitionFoundry/gohfc"
	"github.com/s7techlab/hlf-sdk-go/api"
)

// NewMSPIdentityGOHfc converts gohfc.Identity to api.Identity
func NewMSPIdentityGOHfc(identity *gohfc.Identity) (api.Identity, error) {
	var pemKey []byte

	switch key := identity.PrivateKey.(type) {
	case *ecdsa.PrivateKey:
		if keyBytes, err := x509.MarshalECPrivateKey(key); err != nil {
			return nil, err
		} else {
			pemKey = pem.EncodeToMemory(&pem.Block{Type: `EC PRIVATE KEY`, Bytes: keyBytes})
		}
	default:
		return nil, fmt.Errorf("invalid key type: %v", key)
	}

	pemCert := pem.EncodeToMemory(&pem.Block{Type: `CERTIFICATE`, Bytes: identity.Certificate.Raw})

	return NewMSPIdentityBytes(identity.MspId, pemCert, pemKey)
}
