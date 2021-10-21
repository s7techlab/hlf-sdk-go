package util

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/v2/api"
)

const (
	signCertPath = `signcerts`
	keyStorePath = `keystore`
)

func LoadKeyPairFromMSP(mspPath string) ([]byte, []byte, error) {
	_, err := ioutil.ReadDir(mspPath)
	if err != nil {
		return nil, nil, errors.Wrap(err, `failed to read path`)
	}

	// check signcerts/cert.pem
	certBytes, err := ioutil.ReadFile(path.Join(mspPath, signCertPath, `cert.pem`))
	if err != nil {
		return nil, nil, errors.Wrap(err, `failed to read certificate`)
	}

	certBlock, _ := pem.Decode(certBytes)
	if certBlock == nil {
		return nil, nil, api.ErrInvalidPEMStructure
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, errors.Wrap(err, `failed to parse certificate`)
	}

	var pKeyFileName string

	switch pubKey := cert.PublicKey.(type) {
	case *ecdsa.PublicKey:
		h := sha256.New()
		h.Write(elliptic.Marshal(pubKey.Curve, pubKey.X, pubKey.Y))
		pKeyFileName = fmt.Sprintf("%x_sk", h.Sum(nil))
	default:
		return certBytes, nil, errors.Errorf("unknown key format %s, ECDSA expected", cert.PublicKeyAlgorithm)
	}

	// open private key file
	keyBytes, err := ioutil.ReadFile(path.Join(mspPath, keyStorePath, pKeyFileName))
	if err != nil {
		return certBytes, nil, errors.Wrap(err, `failed to ready private key file`)
	}

	return certBytes, keyBytes, nil
}
