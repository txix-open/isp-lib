package endpoint

import (
	"context"
	"reflect"

	"github.com/integration-system/isp-lib/v3/grpc/isp"
	"github.com/pkg/errors"
)

type param struct {
	index   int
	builder ParamBuilder
}

type Handler struct {
	bodyExtractor RequestBodyExtractor
	bodyMapper    ResponseBodyMapper

	func_        reflect.Value
	paramsCount  int
	params       []param
	reqBodyIndex int
	reqBodyType  reflect.Type
}

func NewCaller(
	f interface{},
	bodyExtractor RequestBodyExtractor,
	bodyMapper ResponseBodyMapper,
	paramMappers map[string]ParamMapper,
) (*Handler, error) {
	rt := reflect.TypeOf(f)
	if rt.Kind() != reflect.Func {
		return nil, errors.New("function expected")
	}
	paramsCount := rt.NumIn()
	reqBodyIndex := -1
	func_ := reflect.ValueOf(f)
	var reqBodyType reflect.Type
	params := make([]param, 0)

	for i := 0; i < paramsCount; i++ {
		p := rt.In(i)
		paramType := p.String()
		mapper, ok := paramMappers[paramType]

		if !ok { //maybe it's a request body
			if reqBodyIndex != -1 {
				return nil, errors.Errorf("param mapper not found for type %s", paramType)
			}
			reqBodyIndex = i
			reqBodyType = p
			continue
		}

		// it's a simple param
		params = append(params, param{index: i, builder: mapper.Builder})
	}

	return &Handler{
		bodyExtractor: bodyExtractor,
		bodyMapper:    bodyMapper,
		func_:         func_,
		paramsCount:   paramsCount,
		params:        params,
		reqBodyIndex:  reqBodyIndex,
		reqBodyType:   reqBodyType,
	}, nil
}

func (h *Handler) Handle(ctx context.Context, message *isp.Message) (*isp.Message, error) {
	args := make([]reflect.Value, h.paramsCount)

	if h.reqBodyIndex != -1 {
		value, err := h.bodyExtractor.Extract(ctx, message, h.reqBodyType)
		if err != nil {
			return nil, err
		}
		args[h.reqBodyIndex] = value
		ctx = RequestBodyToContext(ctx, value.Interface())
	}

	for _, p := range h.params {
		value, err := p.builder(ctx, message)
		if err != nil {
			return nil, err
		}
		args[p.index] = reflect.ValueOf(value)
	}

	returned := h.func_.Call(args)

	var result interface{}
	var err error
	for i := 0; i < len(returned); i++ {
		v := returned[i]
		if e, ok := v.Interface().(error); ok && err == nil {
			err = e
			continue
		}
		if result == nil && v.IsValid() {
			result = v.Interface()
			continue
		}
	}
	if err != nil {
		return nil, err
	}

	return h.bodyMapper.Map(result)
}
