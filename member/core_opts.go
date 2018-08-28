package member

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"gopkg.in/yaml.v2"
)

type coreOptions struct {
	peer    api.Peer
	orderer api.Orderer
}

// CoreOpt describes opt which will be applied to coreOptions
type CoreOpt func(c *core) error

// WithPeer allows to use custom instance of peer in core
func WithPeer(peer api.Peer) CoreOpt {
	return func(c *core) error {
		c.localPeer = peer
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

func WithConfigYaml(configPath string) CoreOpt {
	return func(c *core) error {
		configBytes, err := ioutil.ReadFile(configPath)
		if err != nil {
			return errors.Wrap(err, `failed to read config file`)
		}
		if err = yaml.Unmarshal(configBytes, c.config); err != nil {
			return errors.Wrap(err, `failed to parse YAML`)
		}
		return nil
	}
}

func WithConfigRaw(config config.Config) CoreOpt {
	return func(c *core) error {
		c.config = &config
		return nil
	}
}
