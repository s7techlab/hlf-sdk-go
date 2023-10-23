package http

import (
	"fmt"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/s7techlab/hlf-sdk-go/api/config"
)

type Opt func(c *Client) error

// WithYamlConfig allows using YAML config from file
func WithYamlConfig(configPath string) Opt {
	return func(c *Client) error {
		if configBytes, err := os.ReadFile(configPath); err != nil {
			return fmt.Errorf(`read file=%s: %w`, configPath, err)
		} else {
			c.config = new(config.CAConfig)
			if err = yaml.Unmarshal(configBytes, c.config); err != nil {
				return fmt.Errorf(`unmarshal YAML config: %w`, err)
			}
		}
		return nil
	}
}

func WithBytesConfig(configBytes []byte) Opt {
	return func(c *Client) error {
		if err := yaml.Unmarshal(configBytes, c.config); err != nil {
			return fmt.Errorf(`unmarshal YAML config: %w`, err)
		}
		return nil
	}
}

func WithRawConfig(conf *config.CAConfig) Opt {
	return func(c *Client) error {
		c.config = conf
		return nil
	}
}

func WithHTTPClient(client *http.Client) Opt {
	return func(c *Client) error {
		c.client = client
		return nil
	}
}
