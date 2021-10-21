package ca

import (
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/v2/api/config"
	"gopkg.in/yaml.v2"
)

type opt func(c *core) error

// WithYamlConfig allows to use YAML config from file
func WithYamlConfig(configPath string) opt {
	return func(c *core) error {
		if configBytes, err := ioutil.ReadFile(configPath); err != nil {
			return errors.Wrap(err, `failed to read file contents`)
		} else {
			c.config = new(config.CAConfig)
			if err = yaml.Unmarshal(configBytes, c.config); err != nil {
				return errors.Wrap(err, `failed to read YAML content`)
			}
		}
		return nil
	}
}

func WithBytesConfig(configBytes []byte) opt {
	return func(c *core) error {
		if err := yaml.Unmarshal(configBytes, c.config); err != nil {
			return errors.Wrap(err, `failed to read YAML content`)
		}
		return nil
	}
}

func WithRawConfig(conf *config.CAConfig) opt {
	return func(c *core) error {
		c.config = conf
		return nil
	}
}

func WithHTTPClient(client *http.Client) opt {
	return func(c *core) error {
		c.client = client
		return nil
	}
}
