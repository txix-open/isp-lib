package nats

import (
	"github.com/integration-system/go-cmp/cmp"
	_ "github.com/integration-system/isp-lib/atomic"
	"github.com/integration-system/isp-lib/structure"
	"github.com/pkg/errors"
	"sync"
)

type Option func(c *RxNatsClient)

type RxNatsClient struct {
	nc     *NatsClient
	active bool
	lock   sync.RWMutex

	connectionHandler    connectionHandler
	disconnectionHandler disconnectionHandler
	errorHandler         errorHandler

	lastConfiguration structure.NatsConfig
	clientId          string
}

func (c *RxNatsClient) ReceiveConfiguration(cfg structure.NatsConfig) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if !c.active {
		return
	}

	cfg.ClientId = c.clientId
	if !cmp.Equal(c.lastConfiguration, cfg) {
		nc, err := NewNatsStreamingServerClient(
			cfg,
			c.disconnectionHandler,
			c.connectionHandler,
			c.errorHandler,
		)
		if err != nil && c.errorHandler != nil {
			c.errorHandler(errors.WithMessage(err, "connection"))
			return
		}

		if c.nc != nil {
			_ = c.nc.Close()
			nc.resubscribe(c.nc.subs)
		}

		c.nc = nc
		c.lastConfiguration = cfg
		if c.connectionHandler != nil {
			c.connectionHandler(nc, cfg)
		}
	}
}

func (c *RxNatsClient) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.active = false

	if c.nc != nil {
		client := c.nc
		c.nc = nil
		return client.Close()
	}
	return nil
}

func NewRxNatsClient(clientId string, opts ...Option) *RxNatsClient {
	c := &RxNatsClient{active: true, clientId: clientId}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func WhenConnected(handler connectionHandler) Option {
	return func(c *RxNatsClient) {
		c.connectionHandler = handler
	}
}

func WhenDisconnected(handler disconnectionHandler) Option {
	return func(c *RxNatsClient) {
		c.disconnectionHandler = handler
	}
}

func WhenError(handler errorHandler) Option {
	return func(c *RxNatsClient) {
		c.errorHandler = handler
	}
}
