package config

import (
	"errors"

	"github.com/atomyze-ru/hlf-sdk-go/api"
	"github.com/atomyze-ru/hlf-sdk-go/identity"
)

var (
	ErrMSPIDEmpty   = errors.New(`MSP ID is empty`)
	ErrMSPPathEmpty = errors.New(`MSP path is empty`)

	ErrSignerNotFound = errors.New(`signer not found`)
)

type (
	MSP struct {
		ID   string `yaml:"id"`
		Path string `yaml:"path"`
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

	if m.Path == `` {
		return nil, ErrMSPPathEmpty
	}

	return identity.MSPFromPath(m.ID, m.Path, opts...)
}
