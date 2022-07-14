package config

import (
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Crypto    CryptoConfig       `yaml:"crypto"`
	Orderers  []ConnectionConfig `yaml:"orderers"`
	Discovery DiscoveryConfig    `yaml:"discovery"`
	// peer pool for local configuration without gossip discovery
	MSP  []MSPConfig `yaml:"msp"`
	Pool PoolConfig  `yaml:"pool"`
	// if tls is enabled maps TLS certs to discovered peers
	TLSCertsMap []TLSCertsMapperConfig `yaml:"tls_certs_map"`
}

type ConnectionConfig struct {
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

type PoolConfig struct {
	DeliverTimeout Duration `yaml:"deliver_timeout"`
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
	Enabled      bool   `yaml:"enabled"`
	SkipVerify   bool   `yaml:"skip_verify"`
	HostOverride string `yaml:"host_override"`

	// CertContent take precedence over CertPath
	CertContent []byte `yaml:"cert_content"`
	CertPath    string `yaml:"cert_path"`

	// KeyContent take precedence over KeyPath
	KeyContent []byte `yaml:"key_content"`
	KeyPath    string `yaml:"key_path"`

	// CACertContent take precedence over CACertPath
	CACertContent []byte `yaml:"ca_cert_content"`
	CACertPath    string `yaml:"ca_cert_path"`
}

type DiscoveryConfig struct {
	Type string `yaml:"type"`
	// connection to local MSP which will be used for gossip discovery
	Connection *ConnectionConfig `yaml:"connection"`
	// configuration of channels/chaincodes in local(from config) discovery type
	Options DiscoveryConfigOpts `yaml:"options"`
}

// DiscoveryConfigOpts - channel configuration for local config
// contains []DiscoveryChannel
type DiscoveryConfigOpts map[string]interface{}

type DiscoveryChannel struct {
	Name       string               `json:"channel_name" yaml:"name"`
	Chaincodes []DiscoveryChaincode `json:"chaincodes" yaml:"chaincodes"`
	Orderers   []ConnectionConfig   `json:"orderers" yaml:"orderers"`
}

type DiscoveryChaincode struct {
	Name    string `json:"chaincode_name" yaml:"name"`
	Version string `json:"version"`
	Policy  string `json:"policy"`
}

type CryptoConfig struct {
	Type    string          `yaml:"type"`
	Options CryptoSuiteOpts `yaml:"options"`
}

type CryptoSuiteOpts map[string]interface{}

type Duration struct {
	time.Duration
}

type TLSCertsMapperConfig struct {
	Address   string    `yaml:"address"`
	TlsConfig TlsConfig `yaml:"tls"`
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
