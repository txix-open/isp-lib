package config

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestRemoteConfigOverride(t *testing.T) {
	assert := assert.New(t)

	type Config struct {
		A string
		B int
		C bool
		D struct {
			CamelCase string
			B         string
		}
	}

	original := Config{A: "test1", B: 1, C: true, D: struct {
		CamelCase string
		B         string
	}{CamelCase: "test1", B: "test2"}}

	expect := Config{A: "test2", B: 2, C: false, D: struct {
		CamelCase string
		B         string
	}{CamelCase: "test2", B: "test2"}}

	bytes, err := json.Marshal(original)
	if err != nil {
		panic(err)
	}

	assert.Nil(os.Setenv(RemoteConfigEnvPrefix+"_A", "test2"))
	assert.Nil(os.Setenv(RemoteConfigEnvPrefix+"_B", "2"))
	assert.Nil(os.Setenv(RemoteConfigEnvPrefix+"_C", "false"))
	assert.Nil(os.Setenv(RemoteConfigEnvPrefix+"_D.CAMELCASE", "test2"))

	ptr := InitRemoteConfig(&original, string(bytes)).(*Config)
	original = *ptr

	assert.Equal(expect, original)
}
