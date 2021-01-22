package scripts

import (
	"bytes"
	"sync"
	"time"

	"github.com/dop251/goja"
)

type Script struct {
	prog *goja.Program
}

func NewScript(source ...[]byte) (Script, error) {
	prog, err := goja.Compile("script", string(bytes.Join(source, []byte("\n"))), false)
	return Script{prog: prog}, err
}

func NewFuncScript(scriptBody []byte) (Script, error) {
	return NewScript([]byte("(function() {\n"), scriptBody, []byte("\n})();"))
}

type Engine struct {
	pool *sync.Pool
}

func NewEngine() *Engine {
	return &Engine{&sync.Pool{
		New: func() interface{} {
			vm := goja.New()
			return vm
		},
	}}
}

func (m *Engine) Execute(s Script, arg interface{}, opts ...ExecOption) (interface{}, error) {
	vm := m.pool.Get().(*goja.Runtime)

	config := &configOptions{
		arg:           arg,
		scriptTimeout: 2 * time.Second,
	}
	for _, o := range opts {
		o(config)
	}
	config.timer = time.AfterFunc(config.scriptTimeout, func() {
		vm.Interrupt("execution timeout")
	})
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

func castErr(err error) error {
	if exception, ok := err.(*goja.Exception); ok {
		val := exception.Value().Export()
		if castedErr, ok := val.(error); ok {
			return castedErr
		}
	}
	return err
}
