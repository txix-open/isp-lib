package script

import (
	"bytes"
	"sync"
	"time"

	"github.com/dop251/goja"
)

type Script struct {
	prog *goja.Program
}

type ExecOption func(opt *scriptConfigOptions)

type Machine struct {
	*sync.Pool
}

type scriptConfigOptions struct {
	timer         *time.Timer
	logBuf        *bytes.Buffer
	scriptTimeout time.Duration
	data          map[string]interface{}
}

var byteNewLine = []byte("\n")

func initVm() *goja.Runtime {
	vm := goja.New()
	return vm
}

func InitMachine() *Machine {
	return &Machine{&sync.Pool{
		New: func() interface{} {
			return initVm()
		},
	}}
}

func Create(source ...[]byte) (Script, error) {
	prog, err := goja.Compile("script", string(bytes.Join(source, byteNewLine)), false)
	return Script{prog: prog}, err
}

func (m *Machine) Execute(s Script, arg interface{}, opts ...ExecOption) (interface{}, error) {
	vm := m.Get().(*goja.Runtime)
	config := makeConfig(vm, opts...)
	defer func() {
		config.timer.Stop()
		vm.ClearInterrupt()
		m.Put(vm)
	}()

	console := newConsoleLog(config.logBuf)
	vm.Set("console", console)
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

func WithAnotherScriptTimeout(duration time.Duration) ExecOption {
	return func(opt *scriptConfigOptions) {
		opt.scriptTimeout = duration
	}
}

func WithLogging(logBuf *bytes.Buffer) ExecOption {
	return func(opt *scriptConfigOptions) {
		opt.logBuf = logBuf
	}
}

func WithSet(name string, f interface{}) ExecOption {
	return func(opt *scriptConfigOptions) {
		if opt.data == nil {
			opt.data = make(map[string]interface{}, 1)
		}
		opt.data[name] = f
	}
}

func makeConfig(vm *goja.Runtime, opts ...ExecOption) *scriptConfigOptions {
	config := &scriptConfigOptions{scriptTimeout: 2 * time.Second}
	for _, o := range opts {
		o(config)
	}
	config.timer = time.AfterFunc(config.scriptTimeout, func() {
		vm.Interrupt("execution timeout")
	})
	return config
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
