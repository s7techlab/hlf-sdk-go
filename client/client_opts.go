package client

import (
	"fmt"
	"io/ioutil"

	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/client/grpc"
	"github.com/s7techlab/hlf-sdk-go/crypto"
)

// Opt describes opt which will be applied to coreOptions
type Opt func(c *Client) error

func WithSigner(signer msp.SigningIdentity) Opt {
	return func(c *Client) error {
		c.defaultSigner = signer
		c.discoverySigner = signer
		return nil
	}
}

func WithDefaultSigner(signer msp.SigningIdentity) Opt {
	return func(c *Client) error {
		c.defaultSigner = signer
		return nil
	}
}

func WithDiscoverySigner(signer msp.SigningIdentity) Opt {
	return func(c *Client) error {
		c.defaultSigner = signer
		return nil
	}
}

// WithOrderer allows using custom instance of orderer in Client
func WithOrderer(orderer api.Orderer) Opt {
	return func(c *Client) error {
		c.orderer = orderer
		return nil
	}
}

// WithConfigYaml allows passing path to YAML configuration file
func WithConfigYaml(configPath string) Opt {
	return func(c *Client) error {
		configBytes, err := ioutil.ReadFile(configPath)
		if err != nil {
			return errors.Wrap(err, `failed to read config file`)
		}

		c.config = new(config.Config)

		if err = yaml.Unmarshal(configBytes, c.config); err != nil {
			return errors.Wrap(err, `failed to parse YAML`)
		}
		return nil
	}
}

// WithConfigRaw allows passing to Client created config instance
func WithConfigRaw(config config.Config) Opt {
	return func(c *Client) error {
		c.config = &config
		return nil
	}
}

// WithLogger allows to pass custom copy of zap.Logger insteadof logger.DefaultLogger
func WithLogger(log *zap.Logger) Opt {
	return func(c *Client) error {
		c.logger = log.Named(`hlf-sdk-go`)
		return nil
	}
}

// WithPeerPool allows adding custom peer pool
func WithPeerPool(pool api.PeerPool) Opt {
	return func(c *Client) error {
		c.peerPool = pool
		return nil
	}
}

// WithPeers allows to init Client with peers for specified mspID.
func WithPeers(mspID string, peers []config.ConnectionConfig) Opt {
	return func(c *Client) error {
		for _, p := range peers {
			pp, err := NewPeer(c.ctx, p, c.defaultSigner, c.logger)
			if err != nil {
				return fmt.Errorf("create peer: %w", err)
			}
			err = c.peerPool.Add(mspID, pp, StrategyGRPC(grpc.DefaultGrpcCheckPeriod))
			if err != nil {
				return fmt.Errorf("add peer to pool: %w", err)
			}
		}
		return nil
	}
}

// WithCrypto allows to init Client crypto suite.
func WithCrypto(crypto crypto.Suite) Opt {
	return func(c *Client) error {
		c.crypto = crypto
		return nil
	}
}

// WithFabricV2 toggles Client to use fabric version 2.
func WithFabricV2(fabricV2 bool) Opt {
	return func(c *Client) error {
		c.fabricV2 = fabricV2
		return nil
	}
}
