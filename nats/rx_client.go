package nats

import (
	"github.com/integration-system/go-cmp/cmp"
	_ "github.com/integration-system/isp-lib/atomic"
	"github.com/integration-system/isp-lib/structure"
	"github.com/pkg/errors"
	"sync"
)

var (
	ErrNotConnected = errors.New("nats: not connected")
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
}

func (c *RxNatsClient) ReceiveConfiguration(clientId string, cfg structure.NatsConfig) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if !c.active {
		return
	}

	cfg.ClientId = clientId
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

func (c *RxNatsClient) Visit(visitor func(c *NatsClient) error) error {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.nc == nil {
		return ErrNotConnected
	}

	return visitor(c.nc)
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

func NewRxNatsClient(opts ...Option) *RxNatsClient {
	c := &RxNatsClient{active: true}

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
