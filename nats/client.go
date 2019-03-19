package nats

import (
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/structure"
	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
	"math"
	"sync"
	"time"
)

type NatsClient struct {
	stan.Conn
	Addr string

	natsConn    *nats.Conn
	mu          sync.Mutex
	durableSubs []*DurableSub
}

type DurableSub struct {
	stan.Subscription
	handler stan.MsgHandler
	subj    string
}

const (
	Infinity               = math.MaxInt32
	MinAttempts            = 4
	defaultPingIntervalSec = 3
)

var (
	client *NatsClient
	lock   = sync.Mutex{}
)

func (nc *NatsClient) MakeDurableQueueSubscription(subject string, handler stan.MsgHandler) (stan.Subscription, error) {
	if sub, err := nc.QueueSubscribe(subject,
		subject,
		handler,
		stan.DurableName(subject),
		stan.DeliverAllAvailable(),
	); err != nil {
		return nil, err
	} else {
		ds := &DurableSub{sub, handler, subject}
		nc.mu.Lock()
		nc.durableSubs = append(nc.durableSubs, ds)
		nc.mu.Unlock()
		return ds, nil
	}
}

func (nc *NatsClient) Close() error {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	defer nc.natsConn.Close()

	if err := nc.Conn.Close(); err != nil {
		return err
	}
	return nil
}

func (nc *NatsClient) makeReconnectionCallback(natsConfig *structure.NatsConfig) nats.ConnHandler {
	return func(conn *nats.Conn) {
		if stanConn, err := newStanConn(natsConfig, conn); err != nil {
			logger.Errorf("Could not reconnect to nats streaming server %s. error: %v", nc.Addr, err)
		} else {
			logger.Infof("Reconnected to nats streaming server %s", nc.Addr)
			nc.mu.Lock()
			if nc.Conn != nil {
				nc.Conn.Close()
			}
			nc.Conn = stanConn
			for _, ds := range nc.durableSubs {
				ds.Close()
				if sub, err := stanConn.QueueSubscribe(ds.subj, ds.subj, ds.handler, stan.DurableName(ds.subj)); err != nil {
					logger.Error("Could not reestablish subscription", err)
				} else {
					ds.Subscription = sub
				}
			}
			nc.mu.Unlock()
		}
	}
}

func InitDefaultClient(natsConfig *structure.NatsConfig) (nc *NatsClient) {
	lock.Lock()
	defer lock.Unlock()

	c, err := NewNatsStreamingServerClient(natsConfig, nil)
	if err != nil {
		logger.Fatalf("Could not connect to nats streaming server %s. Error: %v", natsConfig.Address, err)
		return
	}

	logger.Infof("Successfully connected to nats streaming server %s", c.Addr)

	client = c

	return client
}

func NewNatsStreamingServerClient(natsConfig *structure.NatsConfig, disconnectionHandler nats.ConnHandler) (*NatsClient, error) {
	addr := natsConfig.Address.GetAddress()
	client := &NatsClient{Addr: addr}
	natsConn, err := nats.Connect(
		addr,
		nats.Name(natsConfig.ClientId),
		nats.MaxReconnects(-1),
		nats.ReconnectBufSize(-1),
		nats.DisconnectHandler(disconnectionHandler),
		nats.ReconnectHandler(client.makeReconnectionCallback(natsConfig)),
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

func CloseDefaultClient() {
	lock.Lock()
	defer lock.Unlock()

	if client != nil {
		if err := client.Close(); err != nil {
			logger.Warn(err)
		}
		client = nil
	}
}

func IsInitialized() bool {
	return client != nil
}

func GetDefaultClient() *NatsClient {
	if !IsInitialized() {
		logger.Fatal("Nats client has not initialized")
	}

	return client
}

func newStanConn(natsConfig *structure.NatsConfig, natsConn *nats.Conn) (stan.Conn, error) {
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
