package config

import (
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Crypto    CryptoConfig     `yaml:"crypto"`
	Orderer   ConnectionConfig `yaml:"orderer"`
	Discovery DiscoveryConfig  `yaml:"discovery"`
	MSP       []MSPConfig      `yaml:"msp"`
}

type ConnectionConfig struct {
	Host    string     `yaml:"host"`
	Tls     TlsConfig  `yaml:"tls"`
	GRPC    GRPCConfig `yaml:"grpc"`
	Timeout Duration   `yaml:"timeout"`
}

type OrdererConfig struct {
	Host    string     `yaml:"host"`
	Tls     TlsConfig  `yaml:"tls"`
	GRPC    GRPCConfig `yaml:"grpc"`
	Timeout Duration   `yaml:"timeout"`
}

type CAConfig struct {
	Crypto CryptoConfig `yaml:"crypto"`
	Host   string       `yaml:"host"`
	Tls    TlsConfig    `yaml:"tls"`
}

type MSPConfig struct {
	Name      string             `yaml:"name"`
	Endorsers []ConnectionConfig `yaml:"endorsers"`
}

type GRPCConfig struct {
	KeepAlive *GRPCKeepAliveConfig `yaml:"keep_alive"`
	Retry     *GRPCRetryConfig     `yaml:"retry"`
}

type GRPCRetryConfig struct {
	// Count for max retries
	Max uint `yaml:"max"`
	// Timeout is used for back-off
	Timeout Duration `yaml:"timeout"`
}

type GRPCKeepAliveConfig struct {
	// See keepalive.ClientParameters.Time, current value in seconds, default: 1 min.
	Time int `yaml:"time" default:"60"`
	// See keepalive.ClientParameters.Timeout, current value in seconds, default: 20 sec.
	Timeout int `yaml:"timeout" default:"20"`
}

type TlsConfig struct {
	Enabled    bool   `yaml:"enabled"`
	CertPath   string `yaml:"cert_path"`
	CACertPath string `yaml:"ca_cert_path"`
}

type DiscoveryConfig struct {
	Type    string              `yaml:"type"`
	Options DiscoveryConfigOpts `yaml:"options"`
}

type DiscoveryConfigOpts map[string]interface{}

type CryptoConfig struct {
	Type    string          `yaml:"type"`
	Options CryptoSuiteOpts `yaml:"options"`
}

type CryptoSuiteOpts map[string]interface{}

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var out string
	var err error

	if err = unmarshal(&out); err != nil {
		return err
	}

	switch {
	case strings.HasSuffix(out, `s`):
		if d.Duration, err = time.ParseDuration(out); err != nil {
			return err
		}
	case strings.HasSuffix(out, `h`):
		if d.Duration, err = time.ParseDuration(out); err != nil {
			return err
		}
	case strings.HasSuffix(out, `m`):
		if d.Duration, err = time.ParseDuration(out); err != nil {
			return err
		}
	default:
		if t, err := strconv.Atoi(out); err != nil {
			return err
		} else {
			d.Duration = time.Millisecond * time.Duration(t)
		}
	}

	return nil
}
