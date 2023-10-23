package crypto

import (
	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/crypto/ecdsa"
)

type Config struct {
	Type    string            `yaml:"type"`
	Options map[string]string `yaml:"options"`
}

var (
	ErrUnknown      = errors.New(`unknown`)
	ErrTypeRequired = errors.New(`type required`)

	DefaultConfig = &Config{
		Type:    ecdsa.Module,
		Options: ecdsa.DefaultOpts,
	}

	DefaultSuite, _ = NewSuiteByConfig(DefaultConfig, false)
)

func NewSuite(name string, opts map[string]string) (Suite, error) {
	switch name {
	case ecdsa.Module:
		return ecdsa.New(opts)

	default:
		return nil, ErrUnknown
	}
}

func NewSuiteByConfig(config *Config, useDefault bool) (Suite, error) {
	if useDefault && config == nil {
		config = DefaultConfig
	}

	if config.Type == `` {
		return nil, ErrTypeRequired
	}

	return NewSuite(config.Type, config.Options)
}
