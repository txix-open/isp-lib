package backend

import (
	proto "github.com/golang/protobuf/ptypes/struct"
	isp "github.com/integration-system/isp-lib/v2/proto/stubs"
	"github.com/integration-system/isp-lib/v2/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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

func validate(ctx RequestCtx, mappedRequestBody interface{}) error {
	return utils.Validate(mappedRequestBody)
}
