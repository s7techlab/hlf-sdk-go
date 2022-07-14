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
	ErrMSPSignKeyPathEmpty  = errors.New(`MSP signkey path is empty`)

	ErrMSPSignCertContentEmpty = errors.New(`MSP signcert content is empty`)
	ErrMSPSignKeyContentEmpty  = errors.New(`MSP signkey content is empty`)

	ErrSignerNotFound = errors.New(`signer not found`)
)

type (
	MSP struct {
		ID   string `yaml:"id"`
		Path string `yaml:"path"`

		// SignCertPath and SignKeyPath take precedence over Path. If they are present, Path will be ignored
		SignCertPath string `yaml:"signcert_path"`
		SignKeyPath  string `yaml:"signkey_path"`

		// if SignCertContent and SignKeyContent are present, Path, SignCertPath and SignKeyPath will be ignored
		SignCertContent []byte `yaml:"signcert_content"`
		SignKeyContent  []byte `yaml:"signkey_content"`
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

	// cert and key contents take precedence over Path and cert and key paths
	if len(m.SignCertContent) != 0 || len(m.SignKeyContent) != 0 {
		if len(m.SignCertContent) == 0 {
			return nil, ErrMSPSignCertContentEmpty
		}

		if len(m.SignKeyContent) == 0 {
			return nil, ErrMSPSignKeyContentEmpty
		}

		opts = append(opts, identity.WithSignCertContent(m.SignCertContent), identity.WithSignKeyContent(m.SignKeyContent))

		return identity.MSPFromPath(m.ID, "", opts...)
	}

	// cert and key paths take precedence over Path
	if m.SignCertPath != `` || m.SignKeyPath != `` {
		if m.SignCertPath == `` {
			return nil, ErrMSPSignCertPathEmpty
		}

		if m.SignKeyPath == `` {
			return nil, ErrMSPSignKeyPathEmpty
		}

		opts = append(opts, identity.WithSignCertPath(m.SignCertPath), identity.WithSignKeyPath(m.SignKeyPath))

		return identity.MSPFromPath(m.ID, "", opts...)
	}

	if m.Path == `` {
		return nil, ErrMSPPathEmpty
	}

	return identity.MSPFromPath(m.ID, m.Path, opts...)
}
