package streaming

import (
	"github.com/integration-system/isp-lib/proto/stubs"
	"github.com/integration-system/isp-lib/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strconv"
)

const endFileSeq = "end&"

var (
	endFile = &isp.Message{Body: &isp.Message_StructBody{
		StructBody: utils.ConvertMapToGrpcStruct(map[string]interface{}{"end": endFileSeq}),
	}}
)

type StreamConsumer func(stream DuplexMessageStream, md metadata.MD) error

type FormData map[string]interface{}

type DuplexMessageStream interface {
	Send(*isp.Message) error
	Recv() (*isp.Message, error)
}

type BeginFile struct {
	FileName      string
	FormDataName  string
	ContentType   string
	ContentLength int64
	FormData      FormData
}

func (bf BeginFile) ToMessage() *isp.Message {
	data := map[string]interface{}{
		"fileName":      bf.FileName,
		"formDataName":  bf.FormDataName,
		"contentType":   bf.ContentType,
		"contentLength": bf.ContentLength,
		"formData":      bf.FormData,
	}
	s := utils.ConvertMapToGrpcStruct(data)
	return &isp.Message{Body: &isp.Message_StructBody{
		StructBody: s,
	}}
}

func (bf *BeginFile) FromMessage(msg *isp.Message) error {
	s := msg.GetStructBody()
	if s == nil {
		return status.Errorf(codes.InvalidArgument, "Could not convert message to BeginFile. Expected a struct")
	}

	fileName, ok := s.Fields["fileName"]
	if ok {
		bf.FileName = fileName.GetStringValue()
	} else {
		return status.Errorf(codes.InvalidArgument, "Could not convert message to BeginFile. Invalid property 'fileName'")
	}

	formDataName, ok := s.Fields["formDataName"]
	if ok {
		bf.FormDataName = formDataName.GetStringValue()
	} else {
		return status.Errorf(codes.InvalidArgument, "Could not convert message to BeginFile. Invalid property 'formDataName'")
	}

	contentType, ok := s.Fields["contentType"]
	if ok {
		bf.ContentType = contentType.GetStringValue()
	} else {
		return status.Errorf(codes.InvalidArgument, "Could not convert message to BeginFile. Invalid property 'contentType'")
	}

	contentLength, ok := s.Fields["contentLength"]
	if ok {
		bf.ContentLength = int64(contentLength.GetNumberValue())
	} else {
		return status.Errorf(codes.InvalidArgument, "Could not convert message to BeginFile. Invalid property 'contentLength'")
	}

	formData, ok := s.Fields["formData"]
	if ok {
		s = formData.GetStructValue()
		if s != nil {
			bf.FormData = utils.ConvertGrpcStructToMap(s.Fields)
		}
	}

	return nil
}

func (fd FormData) GetIntValue(field string) (int64, error) {
	val := int64(0)
	if fd != nil {
		sval, _ := fd[field]
		if i, ok := sval.(float64); ok {
			val = int64(i)
		} else if s, ok := sval.(string); ok {
			ival, err := strconv.Atoi(s)
			if err == nil {
				val = int64(ival)
			}
		}
	}
	if val == 0 {
		return 0, utils.CreateValidationErrorDetailsV2(codes.InvalidArgument, "Bad request",
			field, "Required int value",
		)
	}
	return val, nil
}

func (fd FormData) GetStringValue(field string) (string, error) {
	val := ""
	if fd != nil {
		sval, _ := fd[field]
		if s, ok := sval.(string); ok {
			val = s
		}
	}
	if val == "" {
		return "", utils.CreateValidationErrorDetailsV2(codes.InvalidArgument, "Bad request",
			field, "Required string value",
		)
	}
	return val, nil
}

func FileEnd() *isp.Message {
	return endFile
}

func IsEndOfFile(msg *isp.Message) bool {
	s := msg.GetStructBody()
	if s == nil {
		return false
	}
	val := s.Fields["end"]
	if val == nil {
		return false
	}
	return val.GetStringValue() == endFileSeq
}
