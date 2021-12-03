package cluster

type AddressConfiguration struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

type ModuleInfo struct {
	ModuleName       string
	ModuleVersion    string
	GrpcOuterAddress AddressConfiguration
	Endpoints        []EndpointDescriptor
}

type EndpointDescriptor struct {
	Path             string
	Inner            bool
	UserAuthRequired bool
	Extra            map[string]interface{}
	Handler          interface{} `json:"-"`
}

type ModuleRequirements struct {
	RequiredModules []string
	RequireRoutes   bool
}

type ModuleDependency struct {
	Name     string
	Required bool
}
