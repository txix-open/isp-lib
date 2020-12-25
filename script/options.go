package script

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
