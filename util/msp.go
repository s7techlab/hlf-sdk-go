package util

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/api"
)

const (
	admincertsPath = "admincerts"
	signcertsPath  = "signcerts"
	keystorePath   = "keystore"
	userscertsPath = "users"
)

// LoadKeyPairFromMSP - legacy method. loads ONLY cert from signcerts dir
func LoadKeyPairFromMSP(mspPath string) ([]byte, []byte, error) {
	_, err := ioutil.ReadDir(mspPath)
	if err != nil {
		return nil, nil, errors.Wrap(err, `failed to read path`)
	}

	// check signcerts/cert.pem
	certBytes, err := ioutil.ReadFile(path.Join(mspPath, signcertsPath, `cert.pem`))
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

	pKeyFileName, err := getPrivateKeyFilename(cert)
	if err != nil {
		return certBytes, nil, errors.Wrap(err, `couldn't fetch private key name`)
	}

	// open private key file
	keyBytes, err := ioutil.ReadFile(path.Join(mspPath, keystorePath, pKeyFileName))
	if err != nil {
		return certBytes, nil, errors.Wrap(err, `failed to ready private key file`)
	}

	// match public/private keys
	if _, err := tls.X509KeyPair(certBytes, keyBytes); err != nil {
		return certBytes, nil, errors.Wrap(err, `certificate/key mismatch`)
	}

	return certBytes, keyBytes, nil
}

// LoadKeypairByCert - takes certificate raw bytes and tries to find suitable(by hash) private key
// in 'keystore' dir
func LoadKeypairByCert(mspPath string, certRawBytes []byte) (*x509.Certificate, interface{}, error) {
	certPEMBytes, _ := pem.Decode(certRawBytes)
	if certPEMBytes == nil {
		return nil, nil, errors.Errorf("no pem content for file")
	}

	cert, err := x509.ParseCertificate(certPEMBytes.Bytes)
	if err != nil {
		return nil, nil, errors.Wrap(err, `failed to parse certificate`)
	}

	files, err := ioutil.ReadDir(path.Join(mspPath, keystorePath))
	if err != nil {
		return nil, nil, fmt.Errorf(`failed to read path: %w`, err)
	}
	var (
		keyFound    bool
		keyRawBytes []byte
	)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		keyRawBytes, err = ioutil.ReadFile(path.Join(mspPath, keystorePath, f.Name()))
		if err != nil {
			return nil, nil, fmt.Errorf(`failed to read private key file: %w`, err)
		}
		// match public/private keys
		if _, err = tls.X509KeyPair(certRawBytes, keyRawBytes); err != nil {
			continue
		}
		keyFound = true
		break
	}

	if !keyFound {
		return nil, nil, fmt.Errorf(`couldn't find key for cert №: %v`, cert.SerialNumber.String())
	}

	keyPEMBytes, _ := pem.Decode(keyRawBytes)
	if keyPEMBytes == nil {
		return nil, nil, errors.Errorf("no pem content for file")
	}

	key, err := x509.ParsePKCS8PrivateKey(keyPEMBytes.Bytes)
	if err != nil {
		return nil, nil, errors.Wrap(err, `failed to parse private key`)
	}

	return cert, key, nil
}

func getPrivateKeyFilename(cert *x509.Certificate) (string, error) {
	switch pubKey := cert.PublicKey.(type) {
	case *ecdsa.PublicKey:
		h := sha256.New()
		h.Write(elliptic.Marshal(pubKey.Curve, pubKey.X, pubKey.Y))
		pKeyFileName := fmt.Sprintf("%x_sk", h.Sum(nil))
		return pKeyFileName, nil
	default:
		return "", errors.Errorf("unknown key format %s, ECDSA expected", cert.PublicKeyAlgorithm)
	}
}

// ReadAllFilesFromDir - read all files from dir
func ReadAllFilesFromDir(dir string) ([][]byte, error) {
	log := zap.L().Named(`GetPemMaterialFromDir`)

	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil, err
	}

	content := make([][]byte, 0)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read directory %s", dir)
	}

	for _, f := range files {
		fullName := filepath.Join(dir, f.Name())

		f, err := os.Stat(fullName)
		if err != nil {
			log.Warn("Failed to stat", zap.Any("fullName", fullName), zap.Error(err))
			continue
		}
		if f.IsDir() {
			continue
		}

		log.Debug("Inspecting file", zap.String("fullName", fullName))

		item, err := ioutil.ReadFile(fullName)
		if err != nil {
			return nil, errors.Wrapf(err, "reading from file %s failed", fullName)
		}
		if err != nil {
			log.Warn("Failed reading file", zap.String("fullName", fullName), zap.Error(err))
			continue
		}

		content = append(content, item)
	}

	return content, nil
}
