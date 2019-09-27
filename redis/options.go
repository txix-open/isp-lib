package redis

type initHandler func(c *Client, err error)

type Option func(client *RxClient)

func WithInitHandler(handler initHandler) Option {
	return func(client *RxClient) {
		client.initHandler = handler
	}
}
