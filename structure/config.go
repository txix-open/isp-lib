package structure

type MetricAddress struct {
	AddressConfiguration
	Path string `json:"path"`
}

type MetricConfiguration struct {
	Address                MetricAddress `json:"address"`
	Gc                     bool          `json:"gc"`
	CollectingGCPeriod     int32         `json:"collectingGCPeriod"`
	Memory                 bool          `json:"memory"`
	CollectingMemoryPeriod int32         `json:"collectingMemoryPeriod"`
}
