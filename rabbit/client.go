package rabbit

import (
	"fmt"
	"github.com/streadway/amqp"
	"github.com/integration-system/isp-lib/structure"
	"sync"
	"time"
)

const (
	defaultConcurrentConsumers = 1
	defaultReconnectionTimeout = 1000
)

type ErrHandler func(err error)

type DisconnectionHandler func(err *amqp.Error)

type RabbitConfig struct {
	Address                  structure.AddressConfiguration `valid:"required~Required"`
	Vhost                    string
	User                     string
	Password                 string
	ReconnectionTimeoutMs    int64
	DisconnectionHandler     DisconnectionHandler `json:"-"`
	ReconnectionErrorHandler ErrHandler           `json:"-"`
	SubscriptionErrorHandler ErrHandler           `json:"-"`
}

func (rc RabbitConfig) GetUri() string {
	if rc.User == "" {
		return fmt.Sprintf("amqp://%s/%s", rc.Address.GetAddress(), rc.Vhost)
	} else {
		return fmt.Sprintf("amqp://%s:%s@%s/%s", rc.User, rc.Password, rc.Address.GetAddress(), rc.Vhost)
	}
}

func (rc RabbitConfig) reconnectionTimeout() time.Duration {
	timeout := rc.ReconnectionTimeoutMs
	if timeout <= 0 {
		timeout = defaultReconnectionTimeout
	}
	return time.Duration(timeout) * time.Millisecond
}

type Client struct {
	conn      *amqp.Connection
	publisher *amqp.Channel
	errChan   chan *amqp.Error
	subs      []*Subscription
	lock      sync.RWMutex
	cfg       RabbitConfig
}

func (c *Client) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	c.lock.RLock()
	if c.publisher != nil {
		defer c.lock.RUnlock()

		return client.publisher.Publish(exchange, key, mandatory, immediate, msg)
	} else {
		c.lock.RUnlock()

		c.lock.Lock()
		defer c.lock.Unlock()

		if c.publisher == nil {
			publisher, err := c.conn.Channel()
			if err != nil {
				return err
			}
			client.publisher = publisher
		}

		return client.publisher.Publish(exchange, key, mandatory, immediate, msg)
	}
}

func (c *Client) Declare(dls ...Declaration) error {
	if len(dls) == 0 {
		return nil
	}

	c.lock.RLock()
	defer c.lock.RUnlock()

	channel, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer func() {
		if channel != nil {
			channel.Close()
		}
	}()

	for _, d := range dls {
		if err := d(channel); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) Subscribe(req SubRequest) (*Subscription, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.makeSub(req)
}

func (c *Client) Close(eh ErrHandler) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, s := range c.subs {
		s.Close(eh)
	}

	if c.publisher != nil {
		if err := c.publisher.Close(); err != nil && eh != nil {
			eh(err)
		}
		c.publisher = nil
	}

	if err := c.conn.Close(); err != nil && eh != nil {
		eh(err)
	}
}

func (c *Client) makeSub(req SubRequest) (*Subscription, error) {
	if req.Handler == nil {
		return nil, ErrHandlerRequired
	}
	if req.ConcurrentConsumers <= 0 {
		req.ConcurrentConsumers = defaultConcurrentConsumers
	}
	if req.Name == "" {
		return nil, ErrNameRequired
	}
	for _, s := range c.subs {
		if s.req.Name == req.Name {
			return nil, fmt.Errorf("Sub with name %s already exists", req.Name)
		}
	}

	sub, err := makeSub(req, c)
	if err != nil {
		return nil, err
	} else {
		c.subs = append(c.subs, sub)
		return sub, nil
	}
}

func (c *Client) startConnCheckTask() {
	go func() {
		for err := range c.errChan {
			if c.cfg.DisconnectionHandler != nil {
				c.cfg.DisconnectionHandler(err)
			}

			c.lock.Lock()

			for _, sub := range c.subs {
				sub.lock.Lock()
				sub.active = false
				sub.lock.Unlock()
			}

			if c.publisher != nil {
				c.publisher.Close()
				c.publisher = nil
			}

			connected := false
			for !connected {
				err := c.dial()
				if err != nil {
					if c.cfg.ReconnectionErrorHandler != nil {
						c.cfg.ReconnectionErrorHandler(err)
					}
					time.Sleep(time.Duration(c.cfg.reconnectionTimeout()))
				} else {
					connected = true
					for _, sub := range c.subs {
						newSub, err := makeSub(sub.req, c)
						eh := c.cfg.SubscriptionErrorHandler
						if err != nil {
							if eh != nil {
								eh(err)
							}
						} else {
							sub.lock.Lock()
							for _, h := range sub.channels {
								h.Close() //ensure release all resources
							}
							sub.channels = newSub.channels
							sub.active = true
							sub.lock.Unlock()
						}
					}
				}
			}

			c.lock.Unlock()
		}
	}()
}

func (c *Client) dial() error {
	conn, err := amqp.Dial(c.cfg.GetUri())
	if err != nil {
		return err
	}
	c.conn = conn
	c.errChan = conn.NotifyClose(c.errChan)
	return nil
}

func MakeClient(config RabbitConfig) (*Client, error) {
	errChan := make(chan *amqp.Error)
	c := &Client{
		cfg:     config,
		errChan: errChan,
	}
	if err := c.dial(); err != nil {
		return nil, err
	}
	c.startConnCheckTask()
	return c, nil
}

func makeSub(req SubRequest, c *Client) (*Subscription, error) {
	sub := &Subscription{req: req, owner: c}
	for i := 0; i < req.ConcurrentConsumers; i++ {
		channel, err := c.conn.Channel()
		if err != nil {
			sub.close(nil, false)
			return nil, err
		}
		err = channel.Qos(req.PrefetchSize, 0, false)
		if err != nil {
			channel.Close()
			sub.close(nil, false)
			return nil, err
		}
		sub.channels = append(sub.channels, channel)
	}
	if err := sub.Start(); err != nil {
		sub.close(nil, false)
		return nil, err
	}
	return sub, nil
}
