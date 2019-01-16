package metric

import (
	"encoding/json"
	"github.com/buaazp/fasthttprouter"
	"github.com/rcrowley/go-metrics"
	"github.com/shirou/gopsutil/cpu"
	"github.com/valyala/fasthttp"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/structure"
	"log"
	"sync"
	"time"
)

var (
	router         = fasthttprouter.New()
	lastMetricPath = ""
	statusCheckers = make(map[string]func() interface{}, 0)
	registry       metrics.Registry
	metricServer   *fasthttp.Server
	lock           sync.Mutex
)

const (
	defaultCollectingPeriod = 10
	defaultIpAddress        = "0.0.0.0"
	defaultPath             = "/metrics"
)

func init() {
	registry = metrics.NewRegistry()
	/*_ = metrics.NewRegisteredFunctionalGauge("go routine count", registry, func() int64 {
		return int64(runtime.NumGoroutine())
	})*/
}

func GetRegistry() metrics.Registry {
	return registry
}

func InitHttpServer(metricConfig structure.MetricConfiguration) {
	metricPort := metricConfig.Address.Port
	if metricPort == "" {
		logger.Warn("Port for collecting metrics must be specified")
	}
	metricPath := metricConfig.Address.Path
	if metricPath == "" {
		metricPath = defaultPath
	}
	metricIp := metricConfig.Address.IP
	if metricIp == "" {
		metricIp = defaultIpAddress
	}

	if lastMetricPath != metricPath {
		router.GET(metricPath, handleMetricRequest)
		lastMetricPath = metricPath
	}

	lock.Lock()
	if metricServer != nil {
		err := metricServer.Shutdown()
		if err != nil {
			logger.Warnf("Error when metric server was stopping: %v", err)
		} else {
			metricServer = nil
		}
	}
	lock.Unlock()

	cpu.Info()

	go func() {
		lock.Lock()
		if metricServer != nil {
			return
		}
		metricServer = &fasthttp.Server{
			Handler:      router.Handler,
			WriteTimeout: time.Second * 60,
			ReadTimeout:  time.Second * 60,
		}
		lock.Unlock()

		err := metricServer.ListenAndServe(metricIp + ":" + metricPort)
		if err != nil {
			log.Fatalf("Metric server error: %v", err)
		}
	}()
}

func InitCollectors(newMetricConfig structure.MetricConfiguration,
	oldMetricConfig structure.MetricConfiguration) metrics.Registry {

	if newMetricConfig.Gc != oldMetricConfig.Gc ||
		newMetricConfig.Memory != oldMetricConfig.Memory {
		registry.UnregisterAll()
		if newMetricConfig.Gc {
			collectingGCPeriod := newMetricConfig.CollectingGCPeriod
			if collectingGCPeriod == 0 {
				collectingGCPeriod = defaultCollectingPeriod
			}
			InitGCMetrics(time.Duration(collectingGCPeriod) * time.Second)
		}
		if newMetricConfig.Memory {
			collectingMemoryPeriod := newMetricConfig.CollectingMemoryPeriod
			if collectingMemoryPeriod == 0 {
				collectingMemoryPeriod = defaultCollectingPeriod
			}
			InitMemoryMetrics(time.Duration(collectingMemoryPeriod) * time.Second)
		}
	}
	return registry
}

func InitGCMetrics(duration time.Duration) {
	metrics.RegisterDebugGCStats(registry)
	go metrics.CaptureDebugGCStats(registry, duration)
}

func InitMemoryMetrics(duration time.Duration) {
	metrics.RegisterRuntimeMemStats(registry)
	go metrics.CaptureRuntimeMemStats(registry, duration)
}

func InitHealhcheck(name string, checker func(h metrics.Healthcheck)) {
	hc := metrics.NewHealthcheck(checker)
	_ = registry.Register(name, hc)
}

func InitStatusChecker(name string, checker func() interface{}) {
	statusCheckers[name] = checker
}

func RemoveStatusChecker(name string) {
	delete(statusCheckers, name)
}

func RemoveAllStatusChecker() {
	statusCheckers = make(map[string]func() interface{}, 0)
}

func GerHttpRouter() *fasthttprouter.Router {
	return router
}

func handleMetricRequest(ctx *fasthttp.RequestCtx) {
	registry.RunHealthchecks()
	allMetrics := registry.GetAll()
	if len(statusCheckers) != 0 {
		statuses := map[string]interface{}{}
		for k, v := range statusCheckers {
			statuses[k] = v()
		}
		allMetrics["status"] = statuses
	}
	bytes, _ := json.Marshal(allMetrics)
	ctx.SetBody(bytes)
	ctx.SetStatusCode(fasthttp.StatusOK)
}
