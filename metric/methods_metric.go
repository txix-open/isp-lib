package metric

import (
	"fmt"
	"sync"
	"time"

	"github.com/integration-system/isp-lib/v2/backend"
	"github.com/rcrowley/go-metrics"
)

const (
	defaultSampleSize = 2048
)

type MethodMetrics struct {
	prefix   string
	registry metrics.Registry

	methodHistograms map[string]metrics.Histogram
	methodLock       sync.RWMutex
	errorsCounter    map[string]metrics.Counter
	statusLock       sync.RWMutex
}

func (mm *MethodMetrics) CatchMetric(method string, dur time.Duration, err error) {
	if err != nil {
		mm.getOrRegisterErrorCounter(method).Inc(1)
	} else {
		ms := int64(dur) / 1e6
		mm.getOrRegisterHistogram(method).Update(ms)
	}
}

func (mm *MethodMetrics) getOrRegisterHistogram(method string) metrics.Histogram {
	mm.methodLock.RLock()
	histogram, ok := mm.methodHistograms[method]
	mm.methodLock.RUnlock()
	if ok {
		return histogram
	}

	mm.methodLock.Lock()
	defer mm.methodLock.Unlock()
	if d, ok := mm.methodHistograms[method]; ok {
		return d
	}
	histogram = metrics.GetOrRegisterHistogram(
		fmt.Sprintf("%s.%s", mm.prefix, method),
		mm.registry,
		metrics.NewUniformSample(defaultSampleSize),
	)
	mm.methodHistograms[method] = histogram
	return histogram
}

func (mm *MethodMetrics) getOrRegisterErrorCounter(method string) metrics.Counter {
	mm.statusLock.RLock()
	d, ok := mm.errorsCounter[method]
	mm.statusLock.RUnlock()
	if ok {
		return d
	}

	mm.statusLock.Lock()
	defer mm.statusLock.Unlock()
	if d, ok := mm.errorsCounter[method]; ok {
		return d
	}
	d = metrics.GetOrRegisterCounter(fmt.Sprintf("%s.%s.error", mm.prefix, method), mm.registry)
	mm.errorsCounter[method] = d
	return d
}

func NewMethodMetrics(metricsPrefix string, registry metrics.Registry) *MethodMetrics {
	return &MethodMetrics{
		prefix:           metricsPrefix,
		registry:         registry,
		methodHistograms: make(map[string]metrics.Histogram),
		errorsCounter:    make(map[string]metrics.Counter),
	}
}

func WithMetrics(metrics *MethodMetrics, next backend.Interceptor) backend.Interceptor {
	return func(ctx backend.RequestCtx, proceed func() (interface{}, error)) (interface{}, error) {
		var (
			resp interface{}
			err  error
		)
		now := time.Now()
		if next != nil {
			resp, err = next(ctx, proceed)
		} else {
			resp, err = proceed()
		}
		since := time.Since(now)
		metrics.CatchMetric(ctx.Method(), since, err)
		return resp, err
	}
}
