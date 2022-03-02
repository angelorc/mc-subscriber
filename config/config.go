package config

import (
	"fmt"
	"io/ioutil"

	"github.com/naoina/toml"
)

type Config struct {
	Mailchimp *MailchimpConfig `toml:"mailchimp"`
	Server    *ServerConfig    `toml:"server"`
}

type ServerConfig struct {
	Address string `toml:"address"`
}

type MailchimpConfig struct {
	APIKey string `toml:"api-key"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		return nil, fmt.Errorf("config file not found")
	}

	configFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %s", err)
	}

	var cfg Config

	err = toml.Unmarshal(configFile, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config: %s", err)
	}

	return &cfg, nil
}
