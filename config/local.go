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

type Local struct {
	cfg *viper.Viper

	validator Validator
	file      string
	envPrefix string
}

func NewLocal(opts ...LocalOption) (*Local, error) {
	local := &Local{}
	for _, opt := range opts {
		opt(local)
	}

	viper := viper.New()
	viper.AutomaticEnv()
	if local.envPrefix != "" {
		viper.SetEnvPrefix(local.envPrefix)
	}

	if local.file != "" {
		dir := path.Dir(local.file)
		file := path.Base(local.file)
		viper.AddConfigPath(dir)
		viper.SetConfigFile(file)
		err := viper.ReadInConfig()
		if err != nil {
			return nil, errors.Wrap(err, "read config from file")
		}
	}

	local.cfg = viper

	return local, nil
}

func (l *Local) Set(key string, value interface{}) {
	l.cfg.Set(key, value)
}

func (l Local) Get(key string) interface{} {
	return l.cfg.Get(key)
}

func (l Local) GetString(key string) string {
	return l.cfg.GetString(key)
}

func (l Local) Read(ctx context.Context, ptr interface{}) error {
	err := l.cfg.Unmarshal(&ptr)
	if err != nil {
		return errors.Wrap(err, "unmarshal config")
	}

	if l.validator == nil {
		return nil
	}
	ok, descs := l.validator.Validate(ctx, ptr)
	if ok {
		return nil
	}
	descriptions := make([]string, 0, len(descs))
	for field, err := range descs {
		descriptions = append(descriptions, fmt.Sprintf("%s -> %s", field, err))
	}
	return errors.Errorf("validate config: %s", descriptions)
}
