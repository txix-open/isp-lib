package nats

import (
	"github.com/nats-io/stan.go"
)

type DurableSub struct {
	owner *NatsClient
	stan.Subscription
	handler stan.MsgHandler
	subj    string
}

func (s *DurableSub) Unsubscribe() error {
	if err := s.Subscription.Unsubscribe(); err != nil {
		return err
	}

	s.owner.removeSubWithLock(s.subj)

	return nil
}

func (s *DurableSub) Close() error {
	if err := s.Subscription.Close(); err != nil {
		return err
	}

	s.owner.removeSubWithLock(s.subj)

	return nil
}
