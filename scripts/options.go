package scripts

import (
	"bytes"
	"time"

	"github.com/dop251/goja"
)

type ExecOption func(opt *configOptions)

type configOptions struct {
	scriptTimeout   time.Duration
	timer           *time.Timer
	arg             interface{}
	logBuf          *bytes.Buffer
	data            map[string]interface{}
	fieldNameMapper goja.FieldNameMapper
}

func WithScriptTimeout(duration time.Duration) ExecOption {
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

func WithFieldNameMapper(fieldNameMapper goja.FieldNameMapper) ExecOption {
	return func(opt *configOptions) {
		opt.fieldNameMapper = fieldNameMapper
	}
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
