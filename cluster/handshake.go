package cluster

import (
	"context"
	"net/http"
	"time"

	etpclient "github.com/integration-system/isp-etp-go/v2/client"
	"github.com/integration-system/isp-lib/v3/lb"
	"github.com/pkg/errors"
)

const (
	ErrorConnection = "ERROR_CONNECTION"
	ConfigError     = "ERROR_CONFIG"

	ConfigSendConfigWhenConnected = "CONFIG:SEND_CONFIG_WHEN_CONNECTED"
	ConfigSendConfigChanged       = "CONFIG:SEND_CONFIG_CHANGED"

	ConfigSendRoutesWhenConnected = "CONFIG:SEND_ROUTES_WHEN_CONNECTED"
	ConfigSendRoutesChanged       = "CONFIG:SEND_ROUTES_CHANGED"

	ModuleReady            = "MODULE:READY"
	ModuleSendRequirements = "MODULE:SEND_REQUIREMENTS"
	ModuleUpdateRoutes     = "MODULE:UPDATE_ROUTES"
	ModuleSendConfigSchema = "MODULE:SEND_CONFIG_SCHEMA"

	ModuleConnectionSuffix = "MODULE_CONNECTED"
)

type Handshake struct {
}

func (h Handshake) Do(ctx context.Context, host string) error {
	cli := etpclient.NewClient(etpclient.Config{
		ConnectionReadLimit: 4 * 1024 * 1024,
		HttpClient:          http.DefaultClient,
	})

	configChan := make(chan []byte)
	cli.On(ConfigSendConfigWhenConnected, func(data []byte) {
		configChan <- data
	})

	balancer := lb.NewRoundRobin([]string{"ws://127.0.0.1:7777"})
	host, err := balancer.Next()
	if err != nil {
		return errors.WithMessage(err, "peek config service host")
	}
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	err = cli.Dial(ctx, host)
	if err != nil {
		return errors.WithMessagef(err, "connect to config service %s", host)
	}

	cli.EmitWithAck(ctx, ModuleSendConfigSchema)
}
