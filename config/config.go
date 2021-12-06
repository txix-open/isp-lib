package config

import (
	"context"
	"fmt"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Validator interface {
	Validate(ctx context.Context, value interface{}) (bool, map[string]string)
}

type Config struct {
	cfg       *viper.Viper
	optional  Optional
	mandatory Mandatory

	validator Validator
	file      string
	envPrefix string
}

func New(opts ...Option) (*Config, error) {
	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}

	viper := viper.New()
	viper.AutomaticEnv()
	if cfg.envPrefix != "" {
		viper.SetEnvPrefix(cfg.envPrefix)
	}

	if cfg.file != "" {
		dir := path.Dir(cfg.file)
		file := path.Base(cfg.file)
		viper.AddConfigPath(dir)
		viper.SetConfigFile(file)
		err := viper.ReadInConfig()
		if err != nil {
			return nil, errors.WithMessage(err, "read config from file")
		}
	}

	cfg.cfg = viper

	return cfg, nil
}

func (c *Config) Set(key string, value interface{}) {
	c.cfg.Set(key, value)
}

func (c *Config) Mandatory() Mandatory {
	return c.mandatory
}

func (c *Config) Optional() Optional {
	return c.optional
}

func (c Config) Read(ctx context.Context, ptr interface{}) error {
	err := c.cfg.Unmarshal(&ptr)
	if err != nil {
		return errors.WithMessage(err, "unmarshal config")
	}

	if c.validator == nil {
		return nil
	}
	ok, details := c.validator.Validate(ctx, ptr)
	if ok {
		return nil
	}
	descriptions := make([]string, 0, len(details))
	for field, err := range details {
		descriptions = append(descriptions, fmt.Sprintf("%s -> %s", field, err))
	}
	return errors.Errorf("validate config: %s", descriptions)
}
