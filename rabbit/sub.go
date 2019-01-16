package rabbit

import (
	"errors"
	"fmt"
	"github.com/streadway/amqp"
	"sync"
)

var (
	ErrHandlerRequired = errors.New("Handler is required")
	ErrNameRequired    = errors.New("Name is required")
	ErrAlreadyStarted  = errors.New("Already started")
	ErrAlreadyStopped  = errors.New("Already closed")
	ErrSubIsClosed     = errors.New("Subscription is closed. Try to create a new one")
)

type MsgHandler func(d amqp.Delivery)

type SubRequest struct {
	ConcurrentConsumers int
	Handler             MsgHandler
	Queue               string
	Name                string
	PrefetchSize        int
}

type Subscription struct {
	req      SubRequest
	lock     sync.Mutex
	channels []*amqp.Channel
	owner    *Client
	active   bool
}

/*Unregister all otherwise close all channels and become not reusable*/
func (s *Subscription) Stop() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.active {
		return ErrAlreadyStopped
	}

	if s.channels == nil {
		return ErrSubIsClosed
	}

	for i, c := range s.channels {
		if err := c.Cancel(fmt.Sprintf("%s_%d", s.req.Name, i), false); err != nil {
			s.close(nil, true)
			return err
		}
	}
	s.active = false

	return nil
}

/*Register n concurrent consumers otherwise close all channels and become not reusable*/
func (s *Subscription) Start() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.active {
		return ErrAlreadyStarted
	}

	if s.channels == nil {
		return ErrSubIsClosed
	}

	req := s.req
	for i, c := range s.channels {
		deliveries, err := c.Consume(
			req.Queue,
			fmt.Sprintf("%s_%d", req.Name, i),
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			s.close(nil, true)
			return err
		}
		go func() {
			for d := range deliveries {
				s.req.Handler(d)
			}
		}()
	}
	s.active = true

	return nil
}

func (s *Subscription) Close(eh ErrHandler) {
	s.lock.Lock()

	s.close(eh, true)

	s.lock.Unlock()
}

func (s *Subscription) IsActive() bool {
	return s.active
}

func (s *Subscription) GetRequest() SubRequest {
	return s.req
}

func (s *Subscription) close(eh ErrHandler, lockClient bool) {
	for _, c := range s.channels {
		for err := c.Close(); err != nil && eh != nil; {
			eh(err)
		}
	}
	s.active = false
	s.channels = nil

	subs := make([]*Subscription, 0)
	if lockClient {
		s.owner.lock.Lock()
		defer s.owner.lock.Unlock()
	}
	for _, sub := range s.owner.subs {
		if sub.req.Name != s.req.Name {
			subs = append(subs, sub)
		}
	}
	s.owner.subs = subs
}
