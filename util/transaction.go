package util

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/s7techlab/hlf-sdk-go/crypto"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
)

// NewTxWithNonce generates new transaction id with crypto nonce
func NewTxWithNonce(id msp.SigningIdentity) (string, []byte, error) {
	if nonce, err := crypto.RandomBytes(24); err != nil {
		return ``, nil, errors.Wrap(err, `failed to get nonce`)
	} else {
		if creator, err := id.Serialize(); err != nil {
			return ``, nil, errors.Wrap(err, `failed to get creator`)
		} else {
			return generateTxId(nonce, creator), nonce, nil
		}
	}
}

// generateTxId returns SHA-256 hash of nonce and creator concatenation
func generateTxId(nonce, creator []byte) string {
	f := sha256.New()
	f.Write(append(nonce, creator...))
	return hex.EncodeToString(f.Sum(nil))
}
