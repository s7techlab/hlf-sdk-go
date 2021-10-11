package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/crypto"
	"github.com/s7techlab/hlf-sdk-go/peer"
)

// CoreOpt describes opt which will be applied to coreOptions
type CoreOpt func(c *core) error

// WithContext allows to pass custom context. Otherwise, context.Background is used
func WithContext(ctx context.Context) CoreOpt {
	return func(c *core) error {
		c.ctx = ctx
		return nil
	}
}

// WithOrderer allows to use custom instance of orderer in core
func WithOrderer(orderer api.Orderer) CoreOpt {
	return func(c *core) error {
		c.orderer = orderer
		return nil
	}
}

// WithConfigYaml allows to pass path to YAML configuration file
func WithConfigYaml(configPath string) CoreOpt {
	return func(c *core) error {
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

// WithConfigRaw allows to pass to core created config instance
func WithConfigRaw(config config.Config) CoreOpt {
	return func(c *core) error {
		c.config = &config
		return nil
	}
}

// WithLogger allows to pass custom copy of zap.Logger insteadof logger.DefaultLogger
func WithLogger(log *zap.Logger) CoreOpt {
	return func(c *core) error {
		c.logger = log.Named(`hlf-sdk-go`)
		return nil
	}
}

// WithPeerPool allows to add custom peer pool
func WithPeerPool(pool api.PeerPool) CoreOpt {
	return func(c *core) error {
		c.peerPool = pool
		return nil
	}
}

// WithPeers allows to init core with peers for specified mspID.
func WithPeers(mspID string, peers []config.ConnectionConfig) CoreOpt {
	return func(c *core) error {
		for _, p := range peers {
			pp, err := peer.New(p, c.logger)
			if err != nil {
				return fmt.Errorf("create peer: %w", err)
			}
			err = c.peerPool.Add(mspID, pp, api.StrategyGRPC(5*time.Second))
			if err != nil {
				return fmt.Errorf("add peer to pool: %w", err)
			}
		}
		return nil
	}
}

// WithCrypto allows to init core crypto suite.
func WithCrypto(cc config.CryptoConfig) CoreOpt {
	return func(c *core) error {
		var err error
		c.cs, err = crypto.GetSuite(cc.Type, cc.Options)
		if err != nil {
			return fmt.Errorf("get crypto suite: %w", err)
		}
		return nil
	}
}

// WithFabricV2 toggles core to use fabric version 2.
func WithFabricV2(fabricV2 bool) CoreOpt {
	return func(c *core) error {
		c.fabricV2 = fabricV2
		return nil
	}
}
