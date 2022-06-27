package config

import (
	"errors"

	"github.com/atomyze-ru/hlf-sdk-go/api"
	"github.com/atomyze-ru/hlf-sdk-go/identity"
)

var (
	ErrMSPIDEmpty   = errors.New(`MSP ID is empty`)
	ErrMSPPathEmpty = errors.New(`MSP path is empty`)

	ErrMSPCertPathEmpty = errors.New(`MSP cert path is empty`)
	ErrMSPKeyPathEmpty  = errors.New(`MSP key path is empty`)

	ErrSignerNotFound = errors.New(`signer not found`)
)

type (
	MSP struct {
		ID   string `yaml:"id"`
		Path string `yaml:"path"`

		// CertPath and KeyPath take precedence over Path. If they are, Path will be ignored
		CertPath string `yaml:"cert_path"`
		KeyPath  string `yaml:"key_path"`
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
	if m.CertPath != `` || m.KeyPath != `` {
		if m.CertPath == `` {
			return nil, ErrMSPCertPathEmpty
		}

		if m.KeyPath == `` {
			return nil, ErrMSPKeyPathEmpty
		}

		opts = append(opts, identity.WithSignCertsPath(m.CertPath), identity.WithKeystorePath(m.KeyPath))

		return identity.MSPFromPath(m.ID, "", opts...)
	}

	if m.Path == `` {
		return nil, ErrMSPPathEmpty
	}

	return identity.MSPFromPath(m.ID, m.Path, opts...)
}
