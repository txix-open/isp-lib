package backend

import (
	"errors"
	proto "github.com/golang/protobuf/ptypes/struct"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/proto/stubs"
	"github.com/integration-system/isp-lib/streaming"
	"github.com/integration-system/isp-lib/structure"
	"github.com/integration-system/isp-lib/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"path"
	"reflect"
	"strings"
)

type ErrorHandler func(err error) (interface{}, error)

type Interceptor func(method string, inputData interface{}, md metadata.MD, proceed func() (interface{}, error)) (interface{}, error)

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
	streamConsumers map[string]streaming.StreamConsumer
	errHandler      ErrorHandler
	interceptor     Interceptor
}

func (df *DefaultService) Request(ctx context.Context, msg *isp.Message) (*isp.Message, error) {
	defer func() {
		err := recover()
		if err != nil {
			logger.Errorf("panic: %v", err)
		}
	}()

	handler, md, err := df.getHandler(ctx)
	if err != nil {
		return nil, err
	}

	var dataParam interface{}
	var result interface{}
	dataParam, err = handler.unmarshalAndValidateInputData(msg)
	if err == nil {
		if df.interceptor != nil {
			result, err = df.interceptor(handler.methodName, dataParam, md, func() (interface{}, error) {
				return handler.call(dataParam, md)
			})
		} else {
			result, err = handler.call(dataParam, md)
		}
	}

	if err != nil && df.errHandler != nil {
		result, err = df.errHandler(err)
	}

	if err != nil {
		grpcError, mustLog := ResolveError(err)
		if mustLog {
			logger.Errorf("Method:%s Error:%v", handler.methodName, err)
		}
		return nil, grpcError.Err()
	}

	msg = emptyBody
	if result != nil {
		msg, err = toBytes(result, ctx)
	}
	return msg, err
}

func (df *DefaultService) RequestStream(stream isp.BackendService_RequestStreamServer) error {
	ctx := stream.Context()
	consumer, md, err := df.getStreamHandler(ctx)
	if err != nil {
		return err
	}
	err = consumer(stream, md)
	if err != nil {
		err, mustLog := ResolveError(err)
		if mustLog {
			logger.Error(err)
		}
		return err.Err()
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

func (df *DefaultService) getStreamHandler(ctx context.Context) (streaming.StreamConsumer, metadata.MD, error) {
	method, md, err := getMethodName(ctx)
	if err != nil {
		return nil, nil, err
	}
	handler, present := df.streamConsumers[method]
	if !present {
		return nil, nil, status.Errorf(codes.Unimplemented, "Method [%s] is not implemented", method)
	}
	return handler, md, nil
}

func GetDefaultService(methodPrefix string, handlersStructs ...interface{}) *DefaultService {
	funcs, streams := resolveHandlers(methodPrefix, handlersStructs...)
	return &DefaultService{functions: funcs, streamConsumers: streams}
}

func GetEndpoints(methodPrefix string, handlersStructs ...interface{}) []structure.EndpointConfig {
	endpoints := make([]structure.EndpointConfig, 0)
	/*logger.Infof("Outer grpc address is %s, module_name: %s, version: %s, libVersion: %s",
	addr.GetAddress(), module.ModuleName, module.Version, module.LibVersion)*/
	for _, handlersStruct := range handlersStructs {
		of := reflect.ValueOf(handlersStruct)
		if of.Kind() == reflect.Map {
			for k := range handlersStruct.(map[string]interface{}) {
				endpoints = append(endpoints, structure.EndpointConfig{Path: k, Inner: false})
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

func GetEndpointConfig(methodPrefix string, f reflect.StructField) structure.EndpointConfig {
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
	return structure.EndpointConfig{Path: path.Join("%s/%s/%s", methodPrefix, group, name), Inner: inner}
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

func toBytes(data interface{}, ctx context.Context) (*isp.Message, error) {
	bytes, err := utils.ConvertInterfaceToBytes(data, ctx)
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
		return fun, errors.New("Expected 2 or less params: ([md] [data])")
	}
	fun.dataParamNum = -1
	fun.mdParamNum = -1
	for i := 0; i < inParamsCount; i++ {
		param := fType.In(i)
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

func resolveHandlers(methodPrefix string, handlersStructs ...interface{}) (map[string]function, map[string]streaming.StreamConsumer) {
	functions := make(map[string]function)
	streamHandlers := make(map[string]streaming.StreamConsumer)
	for _, handlersStruct := range handlersStructs {
		of := reflect.ValueOf(handlersStruct)
		if of.Kind() == reflect.Map {
			for k, v := range handlersStruct.(map[string]interface{}) {
				fValue := reflect.ValueOf(v)
				f, err := getFunction(fValue.Type(), fValue)
				if err != nil {
					logger.Warn(err)
					continue
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
						streamHandlers[key] = f
					} else {
						f, err := getFunction(fType, field)
						if err != nil {
							logger.Warn(err)
							continue
						}

						if _, present := functions[key]; present {
							logger.Warnf("Duplicate method handlers for method: %s", key)
						}
						f.methodName = key
						functions[key] = f
					}
				}
			}
		}
	}
	return functions, streamHandlers
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
