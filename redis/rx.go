package redis

import (
	"github.com/integration-system/go-cmp/cmp"
	"github.com/integration-system/isp-lib/structure"
)

type RxClient struct {
	*Client
	open    bool
	lastCfg structure.RedisConfiguration

	initHandler func(c *Client, err error)
}

func (rc *RxClient) ReceiveConfiguration(cfg structure.RedisConfiguration) {
	if !rc.open {
		return
	}
	if !cmp.Equal(rc.lastCfg, cfg) {
		newClient, err := NewClient(cfg)
		if err != nil {
			rc.callInitHandler(newClient, err)
			return
		}

		if rc.Client != nil {
			_ = rc.Client.Close()
			rc.Client = nil
		}

		rc.Client = newClient
		rc.callInitHandler(newClient, err)
	}
}

func (rc *RxClient) Close() error {
	rc.open = false
	if rc.Client != nil {
		client := rc.Client
		rc.Client = nil
		return client.Close()
	}
	return nil
}

func (rc *RxClient) callInitHandler(c *Client, err error) {
	if rc.initHandler != nil {
		rc.initHandler(c, err)
	}
}

func NewRxClient(opts ...Option) *RxClient {
	c := &RxClient{
		open: true,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
