package rc

import (
	"context"
	"fmt"
	"sync"

	"github.com/integration-system/bellows"
	"github.com/integration-system/isp-lib/v3/json"
	"github.com/pkg/errors"
)

type Validator interface {
	Validate(ctx context.Context, value interface{}) (bool, map[string]string)
}

type Config struct {
	prevConfig     json.RawMessage
	overrideConfig json.RawMessage
	validator      Validator
	lock           sync.Locker
}

func New(validator Validator, overrideData json.RawMessage) *Config {
	return &Config{
		prevConfig:     nil,
		overrideConfig: overrideData,
		validator:      validator,
		lock:           &sync.Mutex{},
	}
}

func (c *Config) Upgrade(ctx context.Context, data json.RawMessage, newConfigPtr interface{}, prevConfigPtr interface{}) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	newConfig, err := c.mergeWithOverride(data)
	if err != nil {
		return errors.WithMessage(err, "merge with override new config")
	}

	err = json.Unmarshal(newConfig, newConfigPtr)
	if err != nil {
		return errors.WithMessage(err, "unmarshal new config")
	}

	ok, details := c.validator.Validate(ctx, newConfigPtr)
	if !ok {
		descriptions := make([]string, 0, len(details))
		for field, err := range details {
			descriptions = append(descriptions, fmt.Sprintf("%s -> %s", field, err))
		}
		return errors.Errorf("validate config: %s", descriptions)
	}

	err = json.Unmarshal(c.prevConfig, prevConfigPtr)
	if err != nil {
		return errors.WithMessage(err, "unmarshal previous config")
	}

	c.prevConfig = newConfig

	return nil
}

func (c *Config) mergeWithOverride(data json.RawMessage) ([]byte, error) {
	config := make(map[string]interface{})
	err := json.Unmarshal(data, &config)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal config")
	}

	overrideData := make(map[string]interface{})
	err = json.Unmarshal(c.overrideConfig, &overrideData)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal override data")
	}

	config = bellows.Flatten(config)
	overrideData = bellows.Flatten(overrideData)
	for k, v := range overrideData {
		config[k] = v
	}
	result := bellows.Expand(config)
	config, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.WithMessagef(err, "unexpected type from bellows, expected map, got %T", result)
	}

	data, err = json.Marshal(config)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal config")
	}

	return data, nil
}
