package config

import (
	"io/ioutil"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func NewYamlConfig(configPath string) (*Config, error) {
	if configBytes, err := ioutil.ReadFile(configPath); err != nil {
		return nil, errors.Wrap(err, `failed to read config file`)
	} else {
		var c Config
		if err = yaml.Unmarshal(configBytes, &c); err != nil {
			return nil, errors.Wrap(err, `failed to unmarshal yaml config`)
		}
		return &c, nil
	}
}
