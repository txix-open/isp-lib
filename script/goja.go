package script

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/dop251/goja"
)

type Script struct {
	prog *goja.Program
}

type ExecOption func(opt *scriptExecOptions)

type scriptExecOptions struct {
	logBuf        *bytes.Buffer
	scriptTimeout time.Duration
	data          map[string]interface{}
}

var pool = &sync.Pool{
	New: func() interface{} {
		return initVm()
	},
}

func Execute(s Script, arg interface{}, opts ...ExecOption) (interface{}, error) {
	vm := pool.Get().(*goja.Runtime)

	config := scriptExecOptions{scriptTimeout: 2 * time.Second}
	for _, o := range opts {
		o(&config)
	}
	t := time.AfterFunc(config.scriptTimeout, func() {
		vm.Interrupt("execution timeout")
	})
	defer func() {
		t.Stop()
		vm.ClearInterrupt()
		pool.Put(vm)
	}()

	console := newConsoleLog(config.logBuf)
	vm.Set("console", console)

	//todo: нарушает абстракцию, но не знаю как иначе
	for name, data := range config.data {
		vm.Set(name, data)
	}

	vm.Set("arg", arg)
	res, err := vm.RunProgram(s.prog)
	if err != nil {
		return nil, castErr(err)
	}
	return res.Export(), nil
}

func initVm() *goja.Runtime {
	vm := goja.New()
	return vm
}

//func Create(sharedScript, source []byte) (Script, error) {
//	script := fmt.Sprintf("%s\n%s", sharedScript, source)
//	prog, err := goja.Compile("script", script, false)
//	return Script{prog: prog}, err
//}

func Create(source ...[]byte) (Script, error) {
	var scriptSource string
	for _, s := range source {
		scriptSource = fmt.Sprintf("%s%s", scriptSource, string(s))
	}
	prog, err := goja.Compile("script", scriptSource, false)
	return Script{prog: prog}, err
}

func castErr(err error) error {
	if exception, ok := err.(*goja.Exception); ok {
		val := exception.Value().Export()
		if castedErr, ok := val.(error); ok {
			return castedErr
		}
	}

	return err
}

func WithLogging(logBuf *bytes.Buffer) ExecOption {
	return func(opt *scriptExecOptions) {
		opt.logBuf = logBuf
	}
}

func WithSpecifiedScriptTimeout(duration time.Duration) ExecOption {
	return func(opt *scriptExecOptions) {
		opt.scriptTimeout = duration
	}
}

func WithFunc(name string, f interface{}) ExecOption {
	return func(opt *scriptExecOptions) {
		if opt.data == nil {
			opt.data = make(map[string]interface{}, 1)
		}
		opt.data[name] = f
	}
}

func WithData(name string, data interface{}) ExecOption {
	return func(opt *scriptExecOptions) {
		if opt.data == nil {
			opt.data = make(map[string]interface{}, 1)
		}
		opt.data[name] = data
	}
}
