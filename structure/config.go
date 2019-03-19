package structure

import (
	"encoding/json"
	"fmt"
	"time"
)

type MetricAddress struct {
	AddressConfiguration
	Path string `json:"path"`
}

type MetricConfiguration struct {
	Address                MetricAddress `json:"address" schema:"Metric HTTP server"`
	Gc                     bool          `json:"gc" schema:"Collect garbage collecting statistic"`
	CollectingGCPeriod     int32         `json:"collectingGCPeriod" schema:"GC stat collecting interval,In seconds, default: 10"`
	Memory                 bool          `json:"memory" schema:"Collect memory statistic"`
	CollectingMemoryPeriod int32         `json:"collectingMemoryPeriod" schema:"Memory stat collecting interval,In seconds, default: 10"`
}

type AddressConfiguration struct {
	Port string `json:"port" schema:"Port"`
	IP   string `json:"ip" schema:"Host"`
}

func (addressConfiguration *AddressConfiguration) GetAddress() string {
	return addressConfiguration.IP + ":" + addressConfiguration.Port
}

type RedisConfiguration struct {
	Address   AddressConfiguration `schema:"Address"`
	Password  string               `schema:"Password"`
	DefaultDB int                  `schema:"Default database"`
}

type RabbitConfig struct {
	Address  AddressConfiguration `valid:"required~Required" schema:"Address"`
	Vhost    string               `schema:"Vhost"`
	User     string               `schema:"Username"`
	Password string               `schema:"Password"`
}

func (rc RabbitConfig) GetUri() string {
	if rc.User == "" {
		return fmt.Sprintf("amqp://%s/%s", rc.Address.GetAddress(), rc.Vhost)
	} else {
		return fmt.Sprintf("amqp://%s:%s@%s/%s", rc.User, rc.Password, rc.Address.GetAddress(), rc.Vhost)
	}
}

func (rc RabbitConfig) ReconnectionTimeout() time.Duration {
	/*timeout := rc.ReconnectionTimeoutMs
	if timeout <= 0 {
		timeout = defaultReconnectionTimeout
	}*/
	return 3 * time.Millisecond
}

type DBConfiguration struct {
	Address      string `valid:"required~Required" schema:"Host"`
	Schema       string `valid:"required~Required" schema:"Schema"`
	Database     string `valid:"required~Required" schema:"Database"`
	Port         string `valid:"required~Required" schema:"Port"`
	Username     string `schema:"Username"`
	Password     string `schema:"Password"`
	PoolSize     int    `schema:"Connection pool size,Default is 10 connections per every CPU"`
	CreateSchema bool   `schema:"Enable schema ensuring,Create schema if not exists"`
}

type SyncLoggerConfig struct {
	Enable         bool   `schema:"Enable file logging"`
	Filename       string `json:"filename" yaml:"filename" schema:"File name"`
	MaxSize        int    `json:"-" yaml:"maxsize"`
	MaxAge         int    `json:"-" yaml:"maxage"`
	MaxBackups     int    `json:"-" yaml:"maxbackups"`
	LocalTime      bool   `json:"-" yaml:"localtime"`
	Compress       bool   `json:"compress" yaml:"compress"`
	ImmediateFlush bool   `json:"immediateFlush" yaml:"immediateFlush"`
}

type NatsConfig struct {
	ClusterId       string               `valid:"required~Required" schema:"Cluster ID"`
	Address         AddressConfiguration `valid:"required~Required" schema:"Address"`
	PingAttempts    int                  `schema:"Max ping attempts,When max attempts is reached connection is closed"`
	PintIntervalSec int                  `schema:"Ping interval,In seconds"`
	ClientId        string               `json:"-"`
}

type SocketConfiguration struct {
	Host             string
	Port             string
	Path             string
	Secure           bool
	UrlParams        map[string]string
	ConnectionString string
}

type ElasticConfiguration struct {
	URL         string
	Username    string
	Password    string
	Sniff       *bool
	Healthcheck *bool
	Infolog     string
	Errorlog    string
	Tracelog    string
}

func (ec *ElasticConfiguration) ConvertTo(elasticConfigPtr interface{}) error {
	if bytes, err := json.Marshal(ec); err != nil {
		return err
	} else if err := json.Unmarshal(bytes, elasticConfigPtr); err != nil {
		return err
	}
	return nil
}
