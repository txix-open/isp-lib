package config

import (
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"path"
	"reflect"
	"sync"
	"syscall"

	"github.com/asaskevich/govalidator"
	"github.com/fsnotify/fsnotify"
	"github.com/mohae/deepcopy"
	"github.com/spf13/viper"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/utils"
)

var (
	configInstance       interface{}
	remoteConfigInstance interface{}

	startWatching  = sync.Once{}
	onChangeFunc   interface{}
	errInvalidFunc = errors.New("Expecting func with two pointers to local config type")

	reloadSig = syscall.SIGUSR1
)

func init() {
	ex, _ := os.Executable()

	viper.SetEnvPrefix("isp")
	viper.AutomaticEnv()

	envConfigName := "config"
	configPath := path.Dir(ex)
	if utils.DEV {
		// _, filename, _, _ := runtime.Caller(0)
		configPath = "./conf/"
	}
	if utils.EnvProfileName != "" {
		envConfigName = "config_" + utils.EnvProfileName + ".yml"
	}
	if utils.EnvConfigPath != "" {
		configPath = utils.EnvConfigPath
	}

	viper.SetConfigName(envConfigName)
	viper.AddConfigPath(configPath)
}

func Get() interface{} {
	if configInstance == nil {
		logger.Fatal("ConfigManager isn't init, call first the \"InitConfig\" method")
	}
	return configInstance
}

func GetRemote() interface{} {
	return remoteConfigInstance
}

func InitConfig(configuration interface{}) interface{} {
	return InitConfigV2(configuration, false)
}

func InitConfigV2(configuration interface{}, callOnChangeHandler bool) interface{} {
	configInstance, _ = readConfig(configuration, true)
	validateLocalConfig(configInstance, true)
	if callOnChangeHandler {
		handleConfigChange(configInstance, nil)
	}
	return configInstance
}

func InitRemoteConfig(configuration interface{}, remoteConfig string) interface{} {
	if err := json.Unmarshal([]byte(remoteConfig), configuration); err == nil {
		validateRemoteConfig(configuration)
		remoteConfigInstance = configuration
	} else {
		logger.Fatal("Invalid remote config json format", err)
	}

	return remoteConfigInstance
}

// Example:
// config.OnConfigChange(func(new, old *conf.Configuration) {
//		logger.Info(new, old)
// })
// Callback call after initial loading and after every config files changing.
// On first call new and old configurations are equals
func OnConfigChange(f interface{}) {
	rt := reflect.TypeOf(f)
	if rt.Kind() != reflect.Func || rt.NumIn() != 2 {
		logger.Panic(errInvalidFunc)
	}

	onChangeFunc = f

	startWatching.Do(func() {
		viper.WatchConfig()
		viper.OnConfigChange(func(in fsnotify.Event) {
			reloadConfig()
		})
		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, reloadSig)
		go func() {
			for {
				_, ok := <-sigChan
				if !ok {
					return
				}
				reloadConfig()
			}
		}()
	})
}

func reloadConfig() {
	old := deepcopy.Copy(configInstance)
	newConfig, err := readConfig(configInstance, false)
	if err != nil {
		return
	}
	if err := validateLocalConfig(newConfig, false); err == nil {
		configInstance = newConfig
		handleConfigChange(newConfig, old)
	} else {
		configInstance = old
	}
}

func readConfig(config interface{}, fatal bool) (interface{}, error) {
	if err := viper.ReadInConfig(); err != nil {
		logError(fatal, "Error reading config file, %v", err)
		return nil, err
	} else if err := viper.Unmarshal(config); err != nil {
		logError(fatal, "Unable to decode into struct, %v", err)
		return nil, err
	}
	return config, nil
}

func handleConfigChange(newConfig, oldConfig interface{}) {
	if onChangeFunc == nil {
		return
	}

	rv := reflect.ValueOf(onChangeFunc)
	rt := rv.Type()
	configType := reflect.TypeOf(newConfig).String()
	newCfgType := rt.In(0).String()
	oldCfgType := rt.In(1).String()
	if newCfgType == oldCfgType && newCfgType == configType {
		if oldConfig == nil {
			oldConfig = newConfig
		}
		args := []reflect.Value{reflect.ValueOf(newConfig), reflect.ValueOf(oldConfig)}
		rv.Call(args)
	} else {
		logger.Panic(errInvalidFunc)
	}
}

func validateLocalConfig(config interface{}, fatal bool) error {
	if _, err := govalidator.ValidateStruct(config); err != nil {
		validationErrors := govalidator.ErrorsByField(err)
		logError(fatal, "Local config int't valid. %v", validationErrors)
		return err
	} else {
		return nil
	}
}

func validateRemoteConfig(remoteConfig interface{}) error {
	if _, err := govalidator.ValidateStruct(remoteConfig); err != nil {
		validationErrors := govalidator.ErrorsByField(err)
		logError(true, "Remote config int't valid. %v", validationErrors)
		return err
	} else {
		return nil
	}
}

func logError(fatal bool, fmt string, args ...interface{}) {
	if fatal {
		logger.Fatalf(fmt, args...)
	} else {
		logger.Errorf(fmt, args...)
	}
}
