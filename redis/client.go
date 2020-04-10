package redis

import (
	rd "github.com/go-redis/redis/v7"
	"github.com/integration-system/isp-lib/v2/structure"
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

type Client struct {
	*rd.Client
}

func (c *Client) UseDb(dbIndex DB, f func(p rd.Pipeliner) error) ([]rd.Cmder, error) {
	return c.Pipelined(useDb(dbIndex, f))
}

func (c *Client) UseDbTx(dbIndex DB, f func(p rd.Pipeliner) error) ([]rd.Cmder, error) {
	return c.TxPipelined(useDb(dbIndex, f))
}

func NewClient(cfg structure.RedisConfiguration) (*Client, error) {
	c := rd.NewClient(&rd.Options{
		Addr:     cfg.Address.GetAddress(),
		Password: cfg.Password,
		DB:       cfg.DefaultDB,
	})

	err := c.Ping().Err()
	if err != nil {
		return nil, err
	}

	return &Client{Client: c}, nil
}

func useDb(dbIndex DB, f func(p rd.Pipeliner) error) func(p rd.Pipeliner) error {
	return func(p rd.Pipeliner) error {
		err := p.Select(int(dbIndex)).Err()
		if err != nil {
			return err
		}
		err = f(p)
		if err != nil {
			return err
		}
		return nil
	}
}
