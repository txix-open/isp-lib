package rabbit

import "github.com/streadway/amqp"

type Declaration func(c *amqp.Channel) error

func DeclareExchange(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) Declaration {
	return func(c *amqp.Channel) error {
		return c.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, args)
	}
}

func DeclareQueue(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) Declaration {
	return func(c *amqp.Channel) error {
		_, err := c.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
		return err
	}
}

func DeclateBinding(name, key, exchange string, noWait bool, args amqp.Table) Declaration {
	return func(c *amqp.Channel) error {
		return c.QueueBind(name, key, exchange, noWait, args)
	}
}
