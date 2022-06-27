package config

import (
	"errors"

	"github.com/atomyze-ru/hlf-sdk-go/api"
	"github.com/atomyze-ru/hlf-sdk-go/identity"
)

var (
	ErrMSPIDEmpty   = errors.New(`MSP ID is empty`)
	ErrMSPPathEmpty = errors.New(`MSP path is empty`)

	ErrMSPSignCertPathEmpty = errors.New(`MSP signcert path is empty`)
	ErrMSPKeystorePathEmpty = errors.New(`MSP keystore path is empty`)

	ErrSignerNotFound = errors.New(`signer not found`)
)

type (
	MSP struct {
		ID   string `yaml:"id"`
		Path string `yaml:"path"`

		// SignCertPath and KeystorePath take precedence over Path. If they are, Path will be ignored
		SignCertPath string `yaml:"signcert_path"`
		KeystorePath string `yaml:"keystore_path"`
	}
)

func (m MSP) MustSigner() api.Identity {
	signer, err := m.Signer()
	if err != nil {
		panic(err)
	}

	return signer
}

func (m MSP) Signer() (api.Identity, error) {
	mspConfig, err := m.MSP(identity.WithSkipConfig())
	if err != nil {
		return nil, err
	}

	signer := mspConfig.Signer()
	if signer == nil {
		return nil, ErrSignerNotFound
	}

	return signer, nil
}

func (m MSP) MSP(opts ...identity.MSPOpt) (identity.MSP, error) {
	if m.ID == `` {
		return nil, ErrMSPIDEmpty
	}

	// Cert and key paths take precedence over Path
	if m.SignCertPath != `` || m.KeystorePath != `` {
		if m.SignCertPath == `` {
			return nil, ErrMSPSignCertPathEmpty
		}

		if m.KeystorePath == `` {
			return nil, ErrMSPKeystorePathEmpty
		}

		opts = append(opts, identity.WithSignCertsPath(m.SignCertPath), identity.WithKeystorePath(m.KeystorePath))

		return identity.MSPFromPath(m.ID, "", opts...)
	}

	if m.Path == `` {
		return nil, ErrMSPPathEmpty
	}

	return identity.MSPFromPath(m.ID, m.Path, opts...)
}
