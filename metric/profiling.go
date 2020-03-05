package metric

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	agent "github.com/integration-system/isp-lib/v2/metric/profefe-agent"
	log "github.com/integration-system/isp-log"
	"github.com/integration-system/isp-log/stdcodes"
	"github.com/profefe/profefe/pkg/profile"
	"github.com/valyala/fasthttp"
)

const (
	startProfilingPath = "/profiling/start"
	stopProfilingPath  = "/profiling/stop"
)

const (
	defaultDuration = 10 * time.Second
	defaultInterval = 10 * time.Second
)

var (
	profileAgent *agent.Agent
	profileLock  sync.Mutex
	serviceName  string
)

func InitProfiling(service string) {
	serviceName = service
}

func handleEnableProfilingRequest(ctx *fasthttp.RequestCtx) {
	profileLock.Lock()
	defer profileLock.Unlock()
	if profileAgent != nil {
		createProfilingResponse(ctx, "profiling already enabled", fasthttp.StatusBadRequest)
		return
	}

	args := ctx.QueryArgs()

	collectorAddr := args.Peek("collector_address")
	if collectorAddr == nil {
		createProfilingResponse(ctx, "'collector_address' not specified", fasthttp.StatusBadRequest)
		return
	}

	duration := defaultDuration
	durationArg := args.GetUintOrZero("duration")
	if durationArg != 0 {
		duration = time.Second * time.Duration(durationArg)
	}

	interval := defaultInterval
	intervalArg := args.GetUintOrZero("interval")
	if intervalArg != 0 {
		interval = time.Second * time.Duration(intervalArg)
	}

	typeArg := args.Peek("types")
	if typeArg == nil {
		createProfilingResponse(ctx, "'types' not specified", fasthttp.StatusBadRequest)
		return
	}

	types := strings.Split(string(typeArg), ",")

	opts, err := parseTypes(types, duration)
	if err != nil {
		createProfilingResponse(ctx, fmt.Sprintf("error parsing types: %v", err), fasthttp.StatusBadRequest)
		return
	}

	addr, _ := getOutboundIp()

	opts = append(opts,
		agent.WithLogger(agentLogger),
		agent.WithLabels("host", addr),
		agent.WithTickInterval(interval),
	)

	newAgent := agent.New(string(collectorAddr), serviceName, opts...)
	err = newAgent.Start(context.Background())
	if err != nil {
		createProfilingResponse(ctx, fmt.Sprintf("error starting agent: %v", err), fasthttp.StatusInternalServerError)
		return
	}

	profileAgent = newAgent
	log.Info(stdcodes.ProfilingStart, "starting profiling")
	createProfilingResponse(ctx, "OK", fasthttp.StatusOK)
}

func handleDisableProfilingRequest(ctx *fasthttp.RequestCtx) {
	profileLock.Lock()
	defer profileLock.Unlock()
	if profileAgent != nil {
		_ = profileAgent.Stop()
		profileAgent = nil
		log.Info(stdcodes.ProfilingStop, "stop profiling")
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}

func agentLogger(format string, v ...interface{}) {
	if strings.HasPrefix(format, "send profile") {
		log.Infof(stdcodes.ProfilingSendNewProfile, format, v...)
	} else {
		log.Warnf(stdcodes.ProfilingError, format, v...)
	}
}

func parseTypes(types []string, duration time.Duration) ([]agent.Option, error) {
	opts := make([]agent.Option, 0)

	for _, val := range types {
		if val == "all" {
			opts = append(opts,
				agent.WithCPUProfile(duration),
				agent.WithTraceProfile(duration),
				agent.WithHeapProfile(),
				agent.WithBlockProfile(),
				agent.WithMutexProfile(),
				agent.WithGoroutineProfile(),
				agent.WithThreadcreateProfile(),
			)
			break
		}

		ptype := profile.ProfileType(0)
		_ = ptype.FromString(val)
		if ptype == profile.TypeUnknown {
			return nil, fmt.Errorf("unknown type %s", val)
		}

		switch ptype {
		case profile.TypeCPU:
			opts = append(opts, agent.WithCPUProfile(duration))
		case profile.TypeHeap:
			opts = append(opts, agent.WithHeapProfile())
		case profile.TypeBlock:
			opts = append(opts, agent.WithBlockProfile())
		case profile.TypeMutex:
			opts = append(opts, agent.WithMutexProfile())
		case profile.TypeGoroutine:
			opts = append(opts, agent.WithGoroutineProfile())
		case profile.TypeTrace:
			opts = append(opts, agent.WithTraceProfile(duration))
		case profile.TypeThreadcreate:
			opts = append(opts, agent.WithThreadcreateProfile())
		}
	}

	return opts, nil
}

func createProfilingResponse(ctx *fasthttp.RequestCtx, message string, statusCode int) {
	response := struct {
		Message string
	}{
		Message: message,
	}

	ctx.SetStatusCode(statusCode)
	bytes, _ := json.Marshal(response)
	ctx.SetContentType("application/json")
	ctx.SetBody(bytes)
}

func getOutboundIp() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP.To4().String(), nil
}
