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

type ExecOption func(opt *configOptions)

type Machine struct {
	pool   *sync.Pool
	config configOptions
}

type configOptions struct {
	scriptTimeout   time.Duration
	timer           *time.Timer
	arg             interface{}
	logBuf          *bytes.Buffer
	data            map[string]interface{}
	fieldNameMapper goja.FieldNameMapper
}

func (m *Machine) Execute(s Script, arg interface{}, opts ...ExecOption) (interface{}, error) {
	vm := m.pool.Get().(*goja.Runtime)
	defer m.config.unset(vm)

	config := newConfig(vm, arg, opts...)
	defer func() {
		config.timer.Stop()
		vm.ClearInterrupt()
		m.pool.Put(vm)
	}()
	config.set(vm)
	defer config.unset(vm)

	res, err := vm.RunProgram(s.prog)
	if err != nil {
		return nil, castErr(err)
	}
	return res.Export(), nil
}

func (m *Machine) newVm() *goja.Runtime {
	vm := goja.New()
	m.config.set(vm)
	return vm
}

func (c *configOptions) set(vm *goja.Runtime) {
	vm.Set("arg", c.arg)
	console := newConsoleLog(c.logBuf)
	vm.Set("console", console)
	if c.fieldNameMapper != nil {
		vm.SetFieldNameMapper(c.fieldNameMapper)
	}
	for name, data := range c.data {
		vm.Set(name, data)
	}
}

func (c *configOptions) unset(vm *goja.Runtime) {
	vm.Set("arg", goja.Undefined())
	vm.Set("console", goja.Undefined())
	for name := range c.data {
		vm.Set(name, goja.Undefined())
	}
	vm.SetFieldNameMapper(nil)
}

func NewMachine(opts ...ExecOption) *Machine {
	m := &Machine{}
	m.config = configOptions{}
	for _, o := range opts {
		o(&m.config)
	}
	m.pool = &sync.Pool{
		New: func() interface{} {
			return m.newVm()
		},
	}
	return m
}

func NewScript(source ...[]byte) (Script, error) {
	prog, err := goja.Compile("script", string(bytes.Join(source, []byte("\n"))), false)
	return Script{prog: prog}, err
}

func WithAnotherScriptTimeout(duration time.Duration) ExecOption {
	return func(opt *configOptions) {
		opt.scriptTimeout = duration
	}
}

func WithLogging(logBuf *bytes.Buffer) ExecOption {
	return func(opt *configOptions) {
		opt.logBuf = logBuf
	}
}

func WithSet(name string, f interface{}) ExecOption {
	return func(opt *configOptions) {
		if opt.data == nil {
			opt.data = make(map[string]interface{}, 1)
		}
		opt.data[name] = f
	}
}

func WithSetFieldNameMapper(fieldNameMapper goja.FieldNameMapper) ExecOption {
	return func(opt *configOptions) {
		opt.fieldNameMapper = fieldNameMapper
	}
}

func newConfig(vm *goja.Runtime, arg interface{}, opts ...ExecOption) *configOptions {
	config := &configOptions{
		scriptTimeout: 2 * time.Second,
		arg:           arg,
	}
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
