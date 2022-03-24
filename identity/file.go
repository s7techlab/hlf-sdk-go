package identity

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	ErrNoFilesInDirectory = errors.New(`no files in directory`)
)

// MSP directory contains
// - a folder admincerts to include PEM files each corresponding to an administrator certificate
// - a folder cacerts to include PEM files each corresponding to a root CA’s certificate
// - (optional) a folder intermediatecerts to include PEM files each corresponding to an intermediate CA’s certificate
// - (optional) a file config.yaml to configure the supported Organizational Units and identity classifications (see respective sections below).
// - (optional) a folder crls to include the considered CRLs
// - a folder keystore to include a PEM file with the node’s signing key; we emphasise that currently RSA keys are not supported
// - a folder signcerts to include a PEM file with the node’s X.509 certificate
// - (optional) a folder tlscacerts to include PEM files each corresponding to a TLS root CA’s certificate
// - (optional) a folder tlsintermediatecerts to include PEM files each corresponding to an intermediate TLS CA’s certificate

const (
	MSPAdmincertsPath        = "admincerts"
	MSPCacertsPath           = "cacerts"
	MSPIntermediatecertsPath = "intermediatecerts"
	MSPKeystorePath          = "keystore"
	MSPSigncertsPath         = "signcerts"
	MSPUserscertsPath        = "user"
)

func AdmincertsPath(mspPath string) string {
	return filepath.Join(mspPath, MSPAdmincertsPath)
}

func KeystorePath(mspPath string) string {
	return filepath.Join(mspPath, MSPKeystorePath)
}

func SigncertsPath(mspPath string) string {
	return filepath.Join(mspPath, MSPSigncertsPath)
}

func UsercertsPath(mspPath string) string {
	return filepath.Join(mspPath, MSPUserscertsPath)
}

func CacertsPath(mspPath string) string {
	return filepath.Join(mspPath, MSPCacertsPath)
}

func IntermediatecertsPath(mspPath string) string {
	return filepath.Join(mspPath, MSPIntermediatecertsPath)
}

func readFirstFile(dir string) ([]byte, error) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil, err
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read directory=%s: %w", dir, err)
	}

	for _, f := range files {
		fullName := filepath.Join(dir, f.Name())

		f, err := os.Stat(fullName)
		if err != nil {
			continue
		}
		if f.IsDir() {
			continue
		}

		content, err := ioutil.ReadFile(fullName)
		if err != nil {
			return nil, fmt.Errorf("read from file=%s: %w", fullName, err)
		}

		return content, nil

	}

	return nil, ErrNoFilesInDirectory
}

// readAllFiles - read files from dir
func readFiles(dir string) ([][]byte, error) {
	if _, err := os.Stat(dir); err != nil {
		return nil, err
	}

	contents := make([][]byte, 0)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read directory=%s: %w", dir, err)
	}

	for _, f := range files {
		fullName := filepath.Join(dir, f.Name())

		f, err := os.Stat(fullName)
		if err != nil {
			continue
		}
		if f.IsDir() {
			continue
		}

		content, err := ioutil.ReadFile(fullName)
		if err != nil {
			return nil, fmt.Errorf("read from file=%s: %w", fullName, err)
		}

		contents = append(contents, content)
	}

	return contents, nil
}
