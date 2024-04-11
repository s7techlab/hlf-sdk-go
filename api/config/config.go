package config

import (
	"time"

	"github.com/s7techlab/hlf-sdk-go/crypto"
)

type Config struct {
	Crypto    *crypto.Config     `yaml:"crypto"`
	Orderers  []ConnectionConfig `yaml:"orderers"`
	Discovery DiscoveryConfig    `yaml:"discovery"`
	// peer pool for local configuration without gossip discovery
	MSP  []MSPConfig `yaml:"msp"`
	Pool PoolConfig  `yaml:"pool"`
	// if tls is enabled maps TLS certs to discovered peers
	EndpointsMap []Endpoint `yaml:"endpoints_map"`
}

type ConnectionConfig struct {
	Host    string        `yaml:"host"`
	Tls     TlsConfig     `yaml:"tls"`
	GRPC    GRPCConfig    `yaml:"grpc"`
	Timeout time.Duration `yaml:"timeout"`
}

type CAConfig struct {
	Crypto *crypto.Config `yaml:"crypto"`
	Host   string         `yaml:"host"`
	Tls    TlsConfig      `yaml:"tls"`
}

type PoolConfig struct {
	DeliverTimeout time.Duration `yaml:"deliver_timeout"`
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
	Timeout time.Duration `yaml:"timeout"`
}

type GRPCKeepAliveConfig struct {
	// See keepalive.ClientParameters.Time, current value in seconds, default: 1 min.
	Time int `yaml:"time" default:"60"`
	// See keepalive.ClientParameters.Timeout, current value in seconds, default: 20 sec.
	Timeout int `yaml:"timeout" default:"20"`
}

type TlsConfig struct {
	Enabled    bool `yaml:"enabled"`
	SkipVerify bool `yaml:"skip_verify"`

	// Cert take precedence over CertPath
	Cert     []byte `yaml:"cert"`
	CertPath string `yaml:"cert_path"`

	// Key take precedence over KeyPath
	Key     []byte `yaml:"key"`
	KeyPath string `yaml:"key_path"`

	// CACert take precedence over CACertPath
	CACert     []byte `yaml:"ca_cert"`
	CACertPath string `yaml:"ca_cert_path"`
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

type Endpoint struct {
	Host         string    `yaml:"host"`
	HostOverride string    `yaml:"host_override"`
	TlsConfig    TlsConfig `yaml:"tls"`
}
