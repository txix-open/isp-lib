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

	assert.Nil(os.Setenv(RemoteConfigEnvPrefix+"_A", "test2#{string}"))
	assert.Nil(os.Setenv(RemoteConfigEnvPrefix+"_B", "2#{int}"))
	assert.Nil(os.Setenv(RemoteConfigEnvPrefix+"_C", "false#{bool}"))
	assert.Nil(os.Setenv(RemoteConfigEnvPrefix+"_D.CAMELCASE", "test2#{string}"))

	ptr := InitRemoteConfig(&original, string(bytes)).(*Config)
	original = *ptr

	assert.Equal(expect, original)
}

func TestInitRemoteConfig(t *testing.T) {
	assert := assert.New(t)

	oldConfig, expectedConfig := getFirstConfData()
	remoteConfig, err := json.Marshal(expectedConfig)
	assert.NoError(err)
	assert.Equal(expectedConfig, InitRemoteConfig(oldConfig, string(remoteConfig)))

	secOldConf, secExpConf := getSecondConfData()
	secRemoteConf, err := json.Marshal(secExpConf)
	assert.NoError(err)
	assert.Equal(secExpConf, InitRemoteConfig(secOldConf, string(secRemoteConf)))
}

func getFirstConfData() (oldConfig, newConfig interface{}) {
	type supStructure struct {
		Integer int
		Varchar string
	}
	type config struct {
		Integer               int
		Varchar               string
		SupStructure          supStructure
		MapStringSupStructure map[string]supStructure
	}
	oldConfig = &config{
		Integer:      1,
		Varchar:      "one",
		SupStructure: supStructure{Integer: 1, Varchar: "one"},
		MapStringSupStructure: map[string]supStructure{
			"one":   {Integer: 1, Varchar: "one"},
			"two":   {Integer: 2, Varchar: "two"},
			"three": {Integer: 3, Varchar: "three"},
		},
	}
	newConfig = &config{
		Integer: 2,
		Varchar: "two",
		MapStringSupStructure: map[string]supStructure{
			"two":   {Integer: 4, Varchar: "four"},
			"three": {Integer: 3, Varchar: "three"},
		},
	}
	return oldConfig, newConfig
}

func getSecondConfData() (oldConfig, newConfig interface{}) {
	type config struct {
		Integer int
		Varchar string
	}
	oldConfig = &config{
		Integer: 1,
	}
	newConfig = &config{
		Varchar: "one",
	}
	return oldConfig, newConfig
}
