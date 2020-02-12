package bootstrap

import (
	"context"
	"os"

	"github.com/integration-system/isp-lib/v2/structure"
	"github.com/integration-system/isp-lib/v2/utils"
)

// contains information for config service announcement
type ModuleInfo struct {
	ModuleName       string
	ModuleVersion    string
	GrpcOuterAddress structure.AddressConfiguration
	// Deprecated: use Endpoints instead
	Handlers  []interface{}
	Endpoints []structure.EndpointDescriptor
}

// contains base requirements for module
type ModuleRequirements struct {
	RequiredModules []string
	RequireRoutes   bool
}

func (r ModuleRequirements) IsEmpty() bool {
	return len(r.RequiredModules) == 0 && !r.RequireRoutes
}

// invoked once, returns config service address
type socketConfigProducer func(localConfigPtr interface{}) structure.SocketConfiguration

// invoked once before module shutdown
type shutdownHandler func(ctx context.Context, sig os.Signal)

// may invoked many times
type moduleInfoProducer func(localConfigPtr interface{}) ModuleInfo

// receives address list, must returns true if list successfully handled (e.g. established connection)
type addressListConsumer func(list []structure.AddressConfiguration) bool

// receives routes, must returns true if routes successfully handled
type routesConsumer func(routes structure.RoutingConfig) bool

// invoked once, provides object which can send module declaration any time
type declaratorAcquirer func(dec RoutesDeclarator)

type RoutesDeclarator interface {
	DeclareRoutes()
}

//default module declarator, send declaration data with MODULE:UPDATE_ROUTES event
type declarator struct {
	f func(eventType string)
}

func (d *declarator) DeclareRoutes() {
	if d != nil {
		d.f(utils.ModuleUpdateRoutes)
	}
}

// receive from config service when some required module connected/disconnected
type connectEvent struct {
	module      string
	addressList []structure.AddressConfiguration
}

type connectConsumer struct {
	consumer    addressListConsumer
	mustConnect bool
}
