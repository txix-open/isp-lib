package rd

import (
	rd "github.com/go-redis/redis"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/structure"
)

var (
	client *RdClient
)

type DB int

const (
	ApplicationTokenDb DB = iota
	ApplicationPermissionDb
	UserTokenDb
	UserPermissionDb
	DeviceTokenDb
	DevicePermissionDb
)

type RdClient struct {
	*rd.Client
}

func (c *RdClient) UseDb(dbIndex DB, f func(p rd.Pipeliner) error) ([]rd.Cmder, error) {
	return c.Pipelined(useDb(dbIndex, f))
}

func (c *RdClient) UseDbTx(dbIndex DB, f func(p rd.Pipeliner) error) ([]rd.Cmder, error) {
	return c.TxPipelined(useDb(dbIndex, f))
}

type RedisConfiguration struct {
	Address   structure.AddressConfiguration `schema:"Address"`
	Password  string                         `schema:"Password"`
	DefaultDB int                            `schema:"Default database"`
}

func InitClient(cfg RedisConfiguration) *RdClient {

	if client != nil {
		err := client.Close()
		if err != nil {
			logger.Warn("Could not close redis connection", err)
		}
		client = nil
	}

	c := rd.NewClient(&rd.Options{
		Addr:     cfg.Address.GetAddress(),
		Password: cfg.Password,
		DB:       cfg.DefaultDB,
	})

	err := c.Ping().Err()
	if err != nil {
		logger.Fatalf("Could not connect to redis %s", cfg.Address.GetAddress())
	}
	client = &RdClient{c}
	return client
}

func GetClient() *RdClient {
	return client
}

func useDb(dbIndex DB, f func(p rd.Pipeliner) error) func(p rd.Pipeliner) error {
	return func(p rd.Pipeliner) error {
		err := p.Select(int(dbIndex)).Err()
		if err != nil {
			logger.Error("Could not select redis db %d", dbIndex)
			return err
		}
		err = f(p)
		if err != nil {
			logger.Error("Redis error", err)
			return err
		}
		return nil
	}
}
