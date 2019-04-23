package utils

import (
	"github.com/asaskevich/govalidator"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"reflect"
	"regexp"
	"unicode"
)

var uuidRegex = regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")

func IsValidUUID(uuid string) bool {
	return uuidRegex.MatchString(uuid)
}

func WrapError(err error) error {
	st := status.New(codes.Unknown, ServiceError)
	return st.Err()
}

func CreateValidationErrorDetails(errorCode codes.Code, errorMessage string, errors map[string]string) error {

	st := status.New(errorCode, errorMessage)
	if errors != nil {
		var violations = make([]*epb.BadRequest_FieldViolation, len(errors))
		counter := 0
		for k, v := range errors {
			arr := []rune(k)
			arr[0] = unicode.ToLower(arr[0])
			violations[counter] = &epb.BadRequest_FieldViolation{
				Field:       string(arr),
				Description: v,
			}
			counter++
		}
		ds, _ := st.WithDetails(&epb.BadRequest{FieldViolations: violations})
		return ds.Err()
	} else {
		return st.Err()
	}
}

func CreateValidationErrorDetailsV2(errorCode codes.Code, errorMessage string, detailsPairs ...string) error {
	l := len(detailsPairs)
	if l == 0 || l%2 != 0 {
		panic("detailsPairs: expecting pairs of field name and error")
		return nil
	}

	errors := make(map[string]string, l/2)
	for i := 0; i < l-1; i++ {
		errors[detailsPairs[i]] = detailsPairs[i+1]
	}
	return CreateValidationErrorDetails(errorCode, errorMessage, errors)
}

func ValidateV2(value interface{}) error {
	rt := reflect.TypeOf(value)
	val := reflect.ValueOf(value)
	kind := rt.Kind()
	if kind == reflect.Ptr || kind == reflect.Interface {
		rt = rt.Elem()
		kind = rt.Kind()
		val = val.Elem()
	}
	if kind == reflect.Array || kind == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			item := val.Index(i)
			err := Validate(item.Interface())
			if err != nil {
				return err
			}
		}
	} else {
		_, err := govalidator.ValidateStruct(value)
		return err
	}
	return nil
}

func Validate(value interface{}) error {
	err := ValidateV2(value)
	errors := govalidator.ErrorsByField(err)
	if len(errors) != 0 {
		return CreateValidationErrorDetails(codes.InvalidArgument, ValidationError, errors)
	}
	return nil
}
