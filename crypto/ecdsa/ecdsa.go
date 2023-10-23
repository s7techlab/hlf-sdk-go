package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"hash"
	"math/big"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
)

const (
	Module    = `ecdsa`
	curveP256 = `P256`
	curveP384 = `P384`
	curveP512 = `P512`

	hashSHA2256 = `SHA2-256`
	hashSHA2384 = `SHA2-384`
	hashSHA3256 = `SHA3-256`
	hashSHA3384 = `SHA3-384`

	sigSHA256 = `SHA256`
	sigSHA384 = `SHA384`
	sigSHA512 = `SHA512`
)

var (
	DefaultOpts = map[string]string{`curve`: `P256`, `signatureAlgorithm`: `SHA256`, `hash`: `SHA2-256`}

	// precomputed curves half order values for efficiency
	ecCurveHalfOrders = map[elliptic.Curve]*big.Int{
		elliptic.P224(): new(big.Int).Rsh(elliptic.P224().Params().N, 1),
		elliptic.P256(): new(big.Int).Rsh(elliptic.P256().Params().N, 1),
		elliptic.P384(): new(big.Int).Rsh(elliptic.P384().Params().N, 1),
		elliptic.P521(): new(big.Int).Rsh(elliptic.P521().Params().N, 1),
	}

	errUnknownCurve              = errors.New(`unknown elliptic curve`)
	errUnknownHash               = errors.New(`unknown hashing algorithm`)
	errUnknownSignatureAlgorithm = errors.New(`unknown signature algorithm`)

	errInvalidPrivateKey = errors.New(`invalid private key, expected ECDSA`)
	errInvalidPublicKey  = errors.New(`invalid public key, expected ECDSA`)
	errInvalidSignature  = errors.New(`invalid ECDSA signature`)
)

func New(opts map[string]string) (*Suite, error) {
	var options Opts
	var err error

	if err = mapstructure.Decode(opts, &options); err != nil {
		return nil, fmt.Errorf(`decode ECDSA options: %w`, err)
	}

	cs := &Suite{}
	if cs.curve, err = getCurve(options.Curve); err != nil {
		return nil, fmt.Errorf(`elliptic curve: %w`, err)
	}
	if cs.hasher, err = getHasher(options.Hash); err != nil {
		return nil, fmt.Errorf(`hasher: %w`, err)
	}
	if cs.sigAlgorithm, err = getSignatureAlgorithm(options.SignatureAlgorithm); err != nil {
		return nil, errors.Wrap(err, `failed to get signature algorithm`)
	}
	return cs, nil
}

type Opts struct {
	Curve              string
	SignatureAlgorithm string
	Hash               string
}

type Suite struct {
	curve        elliptic.Curve
	hasher       func() hash.Hash
	sigAlgorithm x509.SignatureAlgorithm
}
type ecdsaSignature struct {
	R, S *big.Int
}

func (c *Suite) Sign(msg []byte, key interface{}) ([]byte, error) {
	if privateKey, ok := key.(*ecdsa.PrivateKey); !ok {
		return nil, errInvalidPrivateKey
	} else {
		h := c.Hash(msg)
		R, S, err := ecdsa.Sign(rand.Reader, privateKey, h)
		if err != nil {
			return nil, errors.Wrap(err, `failed to sign message`)
		} else {
			preventMalleability(privateKey, S)
		}

		if signature, err := asn1.Marshal(ecdsaSignature{R, S}); err != nil {
			return nil, errors.Wrap(err, `failed to format asn1 signature`)
		} else {
			return signature, nil
		}
	}
}

func (c *Suite) Verify(publicKey interface{}, msg, sig []byte) error {
	if key, ok := publicKey.(*ecdsa.PublicKey); !ok {
		return errInvalidPublicKey
	} else {
		var signature ecdsaSignature
		if _, err := asn1.Unmarshal(sig, &signature); err != nil {
			return errors.Wrap(err, `failed to unmarshal ECDSA signature`)
		}
		if !ecdsa.Verify(key, c.Hash(msg), signature.R, signature.S) {
			return errInvalidSignature
		}
	}
	return nil
}

func (c *Suite) Hash(data []byte) []byte {
	h := c.hasher()
	h.Write(data)
	return h.Sum(nil)
}

func (c *Suite) NewPrivateKey() (interface{}, error) {
	if key, err := ecdsa.GenerateKey(c.curve, rand.Reader); err != nil {
		return nil, errors.Wrap(err, `failed to generate ECDSA private key`)
	} else {
		return key, nil
	}
}

func (c *Suite) GetSignatureAlgorithm() x509.SignatureAlgorithm {
	return c.sigAlgorithm
}

func getCurve(curveType string) (elliptic.Curve, error) {
	switch curveType {
	case curveP256:
		return elliptic.P256(), nil
	case curveP384:
		return elliptic.P384(), nil
	case curveP512:
		return elliptic.P521(), nil
	}
	return nil, errUnknownCurve
}

func getHasher(hashType string) (func() hash.Hash, error) {
	switch hashType {
	case hashSHA2256:
		return sha256.New, nil
	case hashSHA2384:
		return sha512.New384, nil
	case hashSHA3256:
		return sha3.New256, nil
	case hashSHA3384:
		return sha3.New384, nil
	}
	return nil, errUnknownHash
}

func getSignatureAlgorithm(algorithm string) (x509.SignatureAlgorithm, error) {
	switch algorithm {
	case sigSHA256:
		return x509.ECDSAWithSHA256, nil
	case sigSHA384:
		return x509.ECDSAWithSHA384, nil
	case sigSHA512:
		return x509.ECDSAWithSHA512, nil
	}
	return x509.UnknownSignatureAlgorithm, errUnknownSignatureAlgorithm
}

// from gohfc
func preventMalleability(k *ecdsa.PrivateKey, S *big.Int) {
	halfOrder := ecCurveHalfOrders[k.Curve]
	if S.Cmp(halfOrder) == 1 {
		S.Sub(k.Params().N, S)
	}
}
