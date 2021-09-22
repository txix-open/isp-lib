package backend

import (
	"context"
	"errors"
	"fmt"
	"path"
	"reflect"
	"runtime/debug"
	"strings"

	proto "github.com/golang/protobuf/ptypes/struct"
	"github.com/integration-system/isp-lib/v2/isp"
	"github.com/integration-system/isp-lib/v2/streaming"
	"github.com/integration-system/isp-lib/v2/structure"
	"github.com/integration-system/isp-lib/v2/utils"
	log "github.com/integration-system/isp-log"
	"github.com/integration-system/isp-log/stdcodes"
	pkgerrors "github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ErrorHandler func(err error) (interface{}, error)

type Interceptor func(ctx RequestCtx, proceed func() (interface{}, error)) (interface{}, error)

type PostProcessor func(ctx RequestCtx)

type Validator func(ctx RequestCtx, mappedRequestBody interface{}) error

var (
	metaDataType = reflect.TypeOf(metadata.MD{})
	emptyBody    = &isp.Message{
		Body: &isp.Message_NullBody{
			NullBody: proto.NullValue_NULL_VALUE,
		},
	}
)

type DefaultService struct {
	functions       map[string]function
	streamConsumers map[string]streamFunction
	errHandler      ErrorHandler
	interceptor     Interceptor
	pps             []PostProcessor
	validator       Validator
}

func (df *DefaultService) Request(ctx context.Context, msg *isp.Message) (*isp.Message, error) {
	c := newCtx()
	defer func() {
		err := recover()
		if err != nil {
			log.WithMetadata(log.Metadata{"method": c.method}).
				Errorf(stdcodes.ModuleInternalGrpcServiceError, "recovered panic from request: %v", err)
			debug.PrintStack()
		}

		for _, p := range df.pps {
			p(c)
		}
	}()

	handler, md, err := df.getHandler(ctx)
	if err != nil {
		return nil, err
	}

	c.md = md
	c.method = handler.methodName
	c.requestBody = msg.GetBytesBody()

	var dataParam interface{}
	var result interface{}
	dataParam, err = handler.unmarshalAndValidateInputData(msg, c, df.validator)
	c.err = err
	c.mappedRequest = dataParam
	if err == nil {
		if df.interceptor != nil {
			result, err = df.interceptor(c, func() (interface{}, error) {
				return handler.call(ctx, dataParam, md)
			})
		} else {
			result, err = handler.call(ctx, dataParam, md)
		}
	}

	if err != nil && df.errHandler != nil {
		result, err = df.errHandler(err)
	}

	c.err = err
	c.mappedResponse = result

	if err != nil {
		err = handleError(err, c.method)
	} else {
		msg = emptyBody
		if result != nil {
			msg, err = toBytes(result)

			c.err = err
			if msg != nil {
				c.responseBody = msg.GetBytesBody()
			}
		}
	}

	return msg, err
}

func (df *DefaultService) RequestStream(stream isp.BackendService_RequestStreamServer) error {
	ctx := stream.Context()
	function, md, err := df.getStreamHandler(ctx)
	if err != nil {
		return err
	}
	func() {
		defer func() {
			recovered := recover()
			if recovered != nil {
				err = pkgerrors.WithStack(fmt.Errorf("recovered panic from stream handler: %v", recovered))
			}
		}()
		err = function.consume(stream, md)
	}()
	if err != nil {
		return handleError(err, function.methodName)
	}
	return nil
}

func (df *DefaultService) WithErrorHandler(eh ErrorHandler) *DefaultService {
	df.errHandler = eh
	return df
}

func (df *DefaultService) WithInterceptor(interceptor Interceptor) *DefaultService {
	df.interceptor = interceptor
	return df
}

func (df *DefaultService) WithPostProcessors(pps ...PostProcessor) *DefaultService {
	df.pps = pps
	return df
}

func (df *DefaultService) WithValidator(validator Validator) *DefaultService {
	df.validator = validator
	return df
}

func (df *DefaultService) getHandler(ctx context.Context) (*function, metadata.MD, error) {
	method, md, err := getMethodName(ctx)
	if err != nil {
		return nil, nil, err
	}
	handler, present := df.functions[method]
	if !present {
		if _, present := df.streamConsumers[method]; present {
			return nil, nil, status.Errorf(codes.Unimplemented,
				"Method [%s] accept only binary data. Try add '%s' header",
				method, utils.ExpectFileHeader,
			)
		}
		return nil, nil, status.Errorf(codes.Unimplemented, "Method [%s] is not implemented", method)
	}
	return &handler, md, nil
}

func (df *DefaultService) getStreamHandler(ctx context.Context) (*streamFunction, metadata.MD, error) {
	method, md, err := getMethodName(ctx)
	if err != nil {
		return nil, nil, err
	}
	handler, present := df.streamConsumers[method]
	if !present {
		return nil, nil, status.Errorf(codes.Unimplemented, "Method [%s] is not implemented", method)
	}
	return &handler, md, nil
}

// Deprecated
func GetDefaultService(methodPrefix string, handlersStructs ...interface{}) *DefaultService {
	funcs, streams, err := resolveHandlers(methodPrefix, handlersStructs...)
	if err != nil {
		panic(err)
	}
	return &DefaultService{
		functions:       funcs,
		streamConsumers: streams,
		validator:       validate,
	}
}

func NewDefaultService(descriptors []structure.EndpointDescriptor) *DefaultService {
	funcs, streams, err := resolveHandlersByDescriptors(descriptors)
	if err != nil {
		panic(err)
	}
	return &DefaultService{
		functions:       funcs,
		streamConsumers: streams,
		validator:       validate,
	}
}

// Deprecated
func GetEndpoints(methodPrefix string, handlersStructs ...interface{}) []structure.EndpointDescriptor {
	endpoints := make([]structure.EndpointDescriptor, 0)
	for _, handlersStruct := range handlersStructs {
		of := reflect.ValueOf(handlersStruct)
		if of.Kind() == reflect.Map {
			for k := range handlersStruct.(map[string]interface{}) {
				endpoints = append(endpoints, structure.EndpointDescriptor{Path: k, Inner: false})
			}
		} else {
			t := of.Elem().Type()
			for i := 0; i < t.NumField(); i++ {
				f := t.Field(i)
				if f.Type.Kind() == reflect.Func {
					endpoints = append(endpoints, GetEndpointConfig(methodPrefix, f))
				}
			}
		}
	}

	return endpoints
}

// Deprecated
func GetEndpointConfig(methodPrefix string, f reflect.StructField) structure.EndpointDescriptor {
	name, ok := f.Tag.Lookup("method")
	if !ok {
		name = f.Name
	}
	group, ok := f.Tag.Lookup("group")
	if !ok {
		group = utils.MethodDefaultGroup
	}
	inner := false
	innerString, ok := f.Tag.Lookup("inner")
	if ok && strings.ToLower(innerString) == "true" {
		inner = true
	}
	return structure.EndpointDescriptor{Path: path.Join(methodPrefix, group, name), Inner: inner}
}

func handleError(err error, method string) error {
	grpcError, mustLog := ResolveError(err)
	if mustLog {
		// "%+v" format to expand stacktrace from pkgerrors.WithStack
		log.WithMetadata(log.Metadata{"method": method}).
			Errorf(stdcodes.ModuleInternalGrpcServiceError, "%+v", err)
	} else if utils.DEV {
		log.WithMetadata(log.Metadata{"method": method}).
			Debugf(stdcodes.ModuleInternalGrpcServiceError, "%+v", err)
	}
	return grpcError
}

func readBody(msg *isp.Message, ptr interface{}) error {
	bytes := msg.GetBytesBody()
	if bytes != nil {
		return utils.ConvertBytesToGo(bytes, ptr)
	} else {
		body := ResolveBody(msg)
		return utils.ConvertGrpcToGo(body, ptr)
	}
}

func toBytes(data interface{}) (*isp.Message, error) {
	bytes, err := utils.ConvertGoToBytes(data)
	if err != nil {
		return nil, err
	}
	return &isp.Message{
		Body: &isp.Message_BytesBody{
			BytesBody: bytes,
		},
	}, nil
}

func getFunction(fType reflect.Type, fValue reflect.Value) (function, error) {
	var fun = function{}
	inParamsCount := fType.NumIn()
	if inParamsCount > 2 {
		return fun, errors.New("expected 2 or less params: ([md] [data])")
	}
	fun.dataParamNum = -1
	fun.mdParamNum = -1
	for i := 0; i < inParamsCount; i++ {
		param := fType.In(i)

		switch param.Kind() {
		case reflect.Func:
			return fun, errors.New("unexpected func param: function")
		case reflect.Interface:
			return fun, errors.New("unexpected func param: interface")
		}

		if param.ConvertibleTo(metaDataType) {
			fun.mdParamNum = i
			fun.mdParamType = param
		} else {
			fun.dataParamNum = i
			fun.dataParamType = param
		}
	}
	fun.fun = fValue
	return fun, nil
}

func getStreamConsumer(handler interface{}) streaming.StreamConsumer {
	switch f := handler.(type) {
	case func(streaming.DuplexMessageStream, metadata.MD) error:
		return f
	case streaming.StreamConsumer:
		return f
	case func(streaming.DuplexMessageStream, structure.Isolation) error:
		return func(stream streaming.DuplexMessageStream, md metadata.MD) error {
			return f(stream, structure.Isolation(md))
		}
	default:
		return nil
	}
}

func resolveHandlersByDescriptors(descriptors []structure.EndpointDescriptor) (map[string]function, map[string]streamFunction, error) {
	functions := make(map[string]function)
	streamHandlers := make(map[string]streamFunction)

	for _, descriptor := range descriptors {
		value := reflect.ValueOf(descriptor.Handler)
		if f := getStreamConsumer(descriptor.Handler); f != nil {
			if _, present := streamHandlers[descriptor.Path]; present {
				return nil, nil, fmt.Errorf("duplicate method handlers for method: %s", descriptor.Path)
			}
			streamHandlers[descriptor.Path] = streamFunction{
				methodName: descriptor.Path,
				consume:    f,
			}
		} else {
			f, err := getFunction(value.Type(), value)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid function for method %s: %v", descriptor.Path, err)
			}

			if _, present := functions[descriptor.Path]; present {
				return nil, nil, fmt.Errorf("duplicate method handlers for method: %s", descriptor.Path)
			}
			f.methodName = descriptor.Path
			functions[descriptor.Path] = f
		}
	}

	return functions, streamHandlers, nil
}

func resolveHandlers(methodPrefix string, handlersStructs ...interface{}) (map[string]function, map[string]streamFunction, error) {
	functions := make(map[string]function)
	streamHandlers := make(map[string]streamFunction)
	for _, handlersStruct := range handlersStructs {
		of := reflect.ValueOf(handlersStruct)
		if of.Kind() == reflect.Map {
			for k, v := range handlersStruct.(map[string]interface{}) {
				fValue := reflect.ValueOf(v)
				f, err := getFunction(fValue.Type(), fValue)
				if err != nil {
					return nil, nil, err
				}
				functions[k] = f
			}
		} else {
			val := of.Elem()
			t := val.Type()
			for i := 0; i < val.NumField(); i++ {
				field := val.Field(i)
				fType := field.Type()
				if fType.Kind() == reflect.Func {
					config := GetEndpointConfig(methodPrefix, t.Field(i))
					key := config.Path
					if f, ok := field.Interface().(streaming.StreamConsumer); ok {
						streamHandlers[key] = streamFunction{
							methodName: key,
							consume:    f,
						}
					} else {
						f, err := getFunction(fType, field)
						if err != nil {
							return nil, nil, err
						}

						if _, present := functions[key]; present {
							return nil, nil, fmt.Errorf("duplicate method handlers for method: %s", key)

						}
						f.methodName = key
						functions[key] = f
					}
				}
			}
		}
	}
	return functions, streamHandlers, nil
}

func getMethodName(ctx context.Context) (string, metadata.MD, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", nil, status.Errorf(codes.DataLoss, "Metadata [%s] is required", utils.ProxyMethodNameHeader)
	}
	method, ok := md[utils.ProxyMethodNameHeader]
	if !ok {
		return "", nil, status.Errorf(codes.DataLoss, "Metadata [%s] is required", utils.ProxyMethodNameHeader)
	}
	return method[0], md, nil
}
