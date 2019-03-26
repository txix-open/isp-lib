package nats

import (
	"github.com/integration-system/isp-lib/structure"
	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
	"github.com/pkg/errors"
	"math"
	"sync"
	"time"
)

const (
	Infinity               = math.MaxInt32
	MinAttempts            = 4
	defaultPingIntervalSec = 3
)

type connectionHandler func(c *NatsClient, cfg structure.NatsConfig)

type disconnectionHandler func(cfg structure.NatsConfig)

type errorHandler func(err error)

type NatsClient struct {
	stan.Conn
	cfg                  structure.NatsConfig
	connectionHandler    connectionHandler
	disconnectionHandler disconnectionHandler
	errorHandler         errorHandler

	natsConn *nats.Conn
	lock     sync.Mutex
	subs     []*DurableSub
}

func (c *NatsClient) MakeDurableQueueSubscription(subject string, handler stan.MsgHandler) (stan.Subscription, error) {
	if sub, err := c.QueueSubscribe(subject,
		subject,
		handler,
		stan.DurableName(subject),
	); err != nil {
		return nil, err
	} else {
		ds := &DurableSub{owner: c, Subscription: sub, handler: handler, subj: subject}
		c.lock.Lock()
		c.subs = append(c.subs, ds)
		c.lock.Unlock()
		return ds, nil
	}
}

func (c *NatsClient) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	defer c.natsConn.Close()

	if err := c.Conn.Close(); err != nil {
		return err
	}
	return nil
}

func (c *NatsClient) removeSubWithLock(subId string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if len(c.subs) == 0 {
		return
	}

	newSubs := make([]*DurableSub, 0, len(c.subs)-1)
	for _, sub := range c.subs {
		if sub.subj != subId {
			newSubs = append(newSubs, sub)
		}
	}
	c.subs = newSubs
}

func (c *NatsClient) resubscribe(subs []*DurableSub) {
	newSubs := make([]*DurableSub, 0, len(subs))
	for _, ds := range subs {
		_ = ds.Subscription.Close()
		if sub, err := c.Conn.QueueSubscribe(ds.subj, ds.subj, ds.handler, stan.DurableName(ds.subj)); err != nil {
			if c.errorHandler != nil {
				c.errorHandler(errors.WithMessagef(err, "resubscribe to '%s'", ds.subj))
			}
		} else {
			ds.owner = c
			ds.Subscription = sub
			newSubs = append(newSubs, ds)
		}
	}
	c.subs = newSubs
}

func (c *NatsClient) makeReconnectionCallback() nats.ConnHandler {
	return func(conn *nats.Conn) {
		if stanConn, err := newStanConn(c.cfg, conn); err != nil {
			if c.errorHandler != nil {
				c.errorHandler(errors.WithMessage(err, "reconnect"))
			}
		} else {
			if c.connectionHandler != nil {
				c.connectionHandler(c, c.cfg)
			}

			c.lock.Lock()

			if c.Conn != nil {
				_ = c.Conn.Close()
			}
			c.Conn = stanConn
			c.resubscribe(c.subs)

			c.lock.Unlock()
		}
	}
}

func NewNatsStreamingServerClient(
	natsConfig structure.NatsConfig,
	disconnectionHandler disconnectionHandler,
	connectionHandler connectionHandler,
	errorHandler errorHandler,
) (*NatsClient, error) {
	client := &NatsClient{
		cfg:                  natsConfig,
		disconnectionHandler: disconnectionHandler,
		connectionHandler:    connectionHandler,
		errorHandler:         errorHandler,
	}
	natsConn, err := nats.Connect(
		natsConfig.Address.GetAddress(),
		nats.Name(natsConfig.ClientId),
		nats.MaxReconnects(-1),
		nats.ReconnectBufSize(-1),
		nats.DisconnectHandler(func(conn *nats.Conn) {
			if disconnectionHandler != nil {
				disconnectionHandler(natsConfig)
			}
		}),
		nats.ReconnectHandler(client.makeReconnectionCallback()),
	)
	if err != nil {
		return nil, err
	}
	client.natsConn = natsConn

	stanConn, err := newStanConn(natsConfig, natsConn)
	if err != nil {
		client.natsConn.Close()
		return nil, err
	}

	client.Conn = stanConn

	return client, nil
}

func newStanConn(natsConfig structure.NatsConfig, natsConn *nats.Conn) (stan.Conn, error) {
	addr := natsConfig.Address.GetAddress()
	pingInterval := natsConfig.PintIntervalSec
	if pingInterval <= 0 {
		pingInterval = defaultPingIntervalSec
	}
	pingAttempts := natsConfig.PingAttempts
	if pingAttempts <= 0 {
		pingAttempts = MinAttempts
	}
	return stan.Connect(
		natsConfig.ClusterId,
		natsConfig.ClientId,
		stan.NatsConn(natsConn),
		stan.NatsURL(addr),
		stan.Pings(pingInterval, pingAttempts),
		stan.ConnectWait(3*time.Second),
	)
}
