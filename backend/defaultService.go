package backend

import (
	"errors"
	proto "github.com/golang/protobuf/ptypes/struct"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/proto/stubs"
	"github.com/integration-system/isp-lib/streaming"
	"github.com/integration-system/isp-lib/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"reflect"
)

type ErrorHandler func(err error) (interface{}, error)

var (
	metaDataType = reflect.TypeOf(metadata.MD{})
	emptyBody    = &isp.Message{
		Body: &isp.Message_NullBody{
			NullBody: proto.NullValue_NULL_VALUE,
		},
	}
)

type function struct {
	dataParamType reflect.Type
	mdParamType   reflect.Type
	mdParamNum    int
	dataParamNum  int
	fun           reflect.Value
}

type DefaultService struct {
	functions       map[string]function
	streamConsumers map[string]streaming.StreamConsumer
	eh              ErrorHandler
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

	var instance interface{}
	if handler.dataParamType != nil {
		val := reflect.New(handler.dataParamType)
		instance = val.Interface()
		err := readBody(msg, instance)
		if err != nil {
			logger.Error(err)
			return nil, status.Error(codes.InvalidArgument, "Invalid request body")
		}
		err = utils.Validate(instance)
		if err != nil {
			logger.Debug(err)
			return nil, err
		}
	}

	var argCount int
	if handler.mdParamNum > handler.dataParamNum {
		argCount = handler.mdParamNum + 1
	} else {
		argCount = handler.dataParamNum + 1
	}
	args := make([]reflect.Value, argCount)
	if handler.mdParamNum != -1 {
		args[handler.mdParamNum] = reflect.ValueOf(md).Convert(handler.mdParamType)
	}
	if handler.dataParamNum != -1 && instance != nil {
		args[handler.dataParamNum] = reflect.ValueOf(instance).Elem()
	}

	res := handler.fun.Call(args)

	l := len(res)
	var result interface{}
	for i := 0; i < l; i++ {
		v := res[i]
		if e, ok := v.Interface().(error); ok && err == nil {
			err = e
			continue
		}
		if result == nil { // && !v.IsNil()
			result = v.Interface()
			continue
		}
	}

	if err != nil && df.eh != nil {
		result, err = df.eh(err)
	}

	if err != nil {
		grpcError, mustLog := ResolveError(err)
		if mustLog {
			logger.Errorf("Method:%s Error:%v", handler.fun.String(), err)
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
	df.eh = eh
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

func ResolveBody(msg *isp.Message) *proto.Value {
	list := msg.GetListBody()
	st := msg.GetStructBody()
	if list != nil {
		return &proto.Value{Kind: &proto.Value_ListValue{ListValue: list}}
	} else if st != nil {
		return &proto.Value{Kind: &proto.Value_StructValue{StructValue: st}}
	} else {
		return &proto.Value{Kind: &proto.Value_NullValue{NullValue: proto.NullValue_NULL_VALUE}}
	}
}

func WrapBody(value *proto.Value) *isp.Message {
	var result *isp.Message
	switch value.GetKind().(type) {
	case *proto.Value_StructValue:
		result = &isp.Message{
			Body: &isp.Message_StructBody{
				StructBody: value.GetStructValue(),
			},
		}
		break
	case *proto.Value_ListValue:
		result = &isp.Message{
			Body: &isp.Message_ListBody{
				ListBody: value.GetListValue(),
			},
		}
		break
	case *proto.Value_NullValue:
		result = emptyBody
	default:
		logger.Warn("Incorrect result type, expected struct or array or nil. Will return empty response body")
		result = emptyBody
	}
	return result
}

func ResolveError(err error) (s *status.Status, ok bool) {
	s, isGrpcErr := status.FromError(err)
	if isGrpcErr {
		return s, false
	}
	return status.New(codes.Internal, utils.ServiceError), true
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
	if inParamsCount == 2 {
		firstParam := fType.In(0)
		secondParam := fType.In(1)
		if firstParam.ConvertibleTo(metaDataType) {
			fun.mdParamNum = 0
			fun.mdParamType = firstParam
			fun.dataParamNum = 1
			fun.dataParamType = secondParam
		} else if secondParam.ConvertibleTo(metaDataType) {
			fun.mdParamNum = 1
			fun.mdParamType = secondParam
			fun.dataParamNum = 0
			fun.dataParamType = firstParam
		}
	} else if inParamsCount == 1 {
		firstParam := fType.In(0)
		if firstParam.ConvertibleTo(metaDataType) {
			fun.mdParamNum = 0
			fun.mdParamType = firstParam
		} else {
			fun.dataParamNum = 0
			fun.dataParamType = firstParam
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

					fieldName := t.Field(i).Name

					tag := t.Field(i).Tag
					method, present := tag.Lookup("method")
					if !present {
						method = fieldName
					}
					group, ok := tag.Lookup("group")
					if !ok {
						group = utils.MethodDefaultGroup
					} else {
						group = "/" + group + "/"
					}
					key := methodPrefix + group + method
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
