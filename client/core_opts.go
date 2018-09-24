package client

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// CoreOpt describes opt which will be applied to coreOptions
type CoreOpt func(c *core) error

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
		c.logger = log
		return nil
	}
}
