package identity

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/api"
)

var (
	ErrNoPEMContent = errors.New("no pem content")
	ErrKeyNotFound  = errors.New("key not found")
)

func FirstFromPath(mspID string, certDir, keyDir string) (*identity, error) {
	certFile, err := readFirstFile(certDir)
	if err != nil {
		return nil, err
	}

	cert, key, err := KeyPairForCert(certFile, keyDir)
	if err != nil {
		return nil, err
	}

	return New(mspID, cert, key), nil
}

func ListFromPath(mspID string, certDir, keyDir string) ([]api.Identity, error) {
	var identities []api.Identity

	certFiles, err := readFiles(certDir)
	if err != nil {
		return nil, err
	}

	for _, certRaw := range certFiles {
		cert, key, err := KeyPairForCert(certRaw, keyDir)
		if err != nil {
			return nil, err
		}

		identities = append(identities, New(mspID, cert, key))
	}

	return identities, nil
}

func CertificatesFromPath(certDir string) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate

	certFiles, err := readFiles(certDir)
	if err != nil {
		return nil, err
	}

	for _, certRaw := range certFiles {
		cert, err := Certificate(certRaw)
		if err != nil {
			return nil, err
		}

		certs = append(certs, cert)
	}

	return certs, nil
}

func Certificate(certRaw []byte) (*x509.Certificate, error) {
	certPEM, _ := pem.Decode(certRaw)
	if certPEM == nil {
		return nil, ErrNoPEMContent
	}

	cert, err := x509.ParseCertificate(certPEM.Bytes)
	if err != nil {
		return nil, fmt.Errorf(`parse certificate: %w`, err)
	}

	return cert, nil
}

// Key parses raw key btes
func Key(keyRaw []byte) (interface{}, error) {
	keyPEM, _ := pem.Decode(keyRaw)
	if keyPEM == nil {
		return nil, ErrNoPEMContent
	}

	key, err := x509.ParsePKCS8PrivateKey(keyPEM.Bytes)
	if err != nil {
		return nil, fmt.Errorf(`parse key: %w`, err)
	}

	return key, nil
}

func KeyPairForCert(certRaw []byte, keyDir string) (*x509.Certificate, interface{}, error) {
	key, err := KeyForCert(certRaw, keyDir)
	if err != nil {
		return nil, nil, err
	}

	cert, err := Certificate(certRaw)
	if err != nil {
		return nil, nil, err
	}

	return cert, key, nil
}

// KeyForCert returns private key for certificate from keyDir
func KeyForCert(certRaw []byte, keyDir string) (interface{}, error) {
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return nil, fmt.Errorf(`read key dir: %w`, err)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		keyRaw, err := ioutil.ReadFile(path.Join(keyDir, f.Name()))
		if err != nil {
			return nil, fmt.Errorf(`read key file: %w`, err)
		}
		// match public/private keys
		if _, err = tls.X509KeyPair(certRaw, keyRaw); err != nil {
			continue
		}

		return Key(keyRaw)
	}

	return nil, ErrKeyNotFound
}
