package structure

type MetricAddress struct {
	AddressConfiguration
	Path string `json:"path"`
}

type MetricConfiguration struct {
	Address                MetricAddress `json:"address" schema:"Metric HTTP server"`
	Gc                     bool          `json:"gc" schema:"Collect garbage collecting statistic"`
	CollectingGCPeriod     int32         `json:"collectingGCPeriod" schema:"GC stat collecting interval,In seconds, default: 10"`
	Memory                 bool          `json:"memory" schema:"Collect memory statistic"`
	CollectingMemoryPeriod int32         `json:"collectingMemoryPeriod" schema:"Memory stat collecting interval,In seconds, default: 10"`
}
