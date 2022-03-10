package util

import (
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
)

const (
	admincertsPath = "admincerts"
	signcertsPath  = "signcerts"
	keystorePath   = "keystore"
	userscertsPath = "users"
)

// LoadKeyPairFromMSP - legacy method. loads ONLY cert from signcerts dir
func LoadKeyPairFromMSP(mspPath string) (*x509.Certificate, interface{}, error) {
	_, err := ioutil.ReadDir(mspPath)
	if err != nil {
		return nil, nil, fmt.Errorf(`read msp dir: %w`, err)
	}

	var (
		certBytes []byte
	)
	// check certificate in a first file of signcerts folder
	files, err := ioutil.ReadDir(path.Join(mspPath, signcertsPath))
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		certBytes, err = ioutil.ReadFile(path.Join(mspPath, signcertsPath, f.Name()))
		if err != nil {
			return nil, nil, fmt.Errorf(`read certificate: %w`, err)
		}
		cert, key, err := LoadKeypairByCert(mspPath, certBytes)
		if err != nil {
			return nil, nil, fmt.Errorf(`read keypair: %w`, err)
		}

		return cert, key, nil
	}

	return nil, nil, fmt.Errorf("coudn't find certificate in %s", path.Join(mspPath, signcertsPath))
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
