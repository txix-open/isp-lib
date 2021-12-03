package cluster

import (
	"context"

	"github.com/integration-system/isp-lib/v3/lb"
)

type Client struct {
	module ModuleInfo
	lb     *lb.RoundRobin
}

func NewClient(module ModuleInfo, hosts []string) *Client {
	return &Client{
		module: module,
		lb:     lb.NewRoundRobin(hosts),
	}
}

func (c *Client) Run(ctx context.Context) error {

}
