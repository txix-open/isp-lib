package utils

import (
	"github.com/golang/protobuf/ptypes/struct"
	"go/types"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	nilValue = &structpb.Value{Kind: &structpb.Value_NullValue{}}
)

type GrpcValueMarshaler interface {
	ToGrpcValue() *structpb.Value
}

func ConvertMapToGrpcStruct(data map[string]interface{}) *structpb.Struct {
	fields := make(map[string]*structpb.Value, len(data))
	for key, value := range data {
		fields[key] = ConvertInterfaceToGrpcStruct(value)
	}
	return &structpb.Struct{Fields: fields}
}

func ConvertMapMapToGrpcStruct(data map[string]map[string]int) *structpb.Struct {
	fields := make(map[string]*structpb.Value, len(data))
	for key, value := range data {
		fieldsSecond := make(map[string]*structpb.Value, len(value))
		for keySecond, valueSecond := range value {
			fieldsSecond[keySecond] = &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(valueSecond)}}
		}
		fields[key] = &structpb.Value{Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{Fields: fieldsSecond}}}
	}
	return &structpb.Struct{Fields: fields}
}

func ConvertInterfaceToGrpcStruct(data interface{}) *structpb.Value {
	if marshaler, ok := data.(GrpcValueMarshaler); ok {
		return marshaler.ToGrpcValue()
	}

	dataType := reflect.TypeOf(data)
	if dataType == nil {
		return nilValue
	}
	dataKind := dataType.Kind()
	pointer := false
	if dataKind == reflect.Ptr {
		pointer = true
		dataType = dataType.Elem()
		dataKind = dataType.Kind()
	}
	val, isNil := getValue(data, pointer)
	if isNil {
		return nilValue
	}

	switch dataKind {
	case reflect.Struct:
		if t, ok := data.(time.Time); ok {
			if t.IsZero() {
				return nilValue
			}
			return &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: t.Format(FullDateFormat)}}
		}

		fieldsCount := dataType.NumField()
		fields := make(map[string]*structpb.Value, fieldsCount)
		for i := 0; i < fieldsCount; i++ {
			field := val.Field(i)
			if field.CanInterface() {
				fieldType := dataType.Field(i)
				if name, accept := GetFieldName(fieldType); accept {
					if fieldType.Anonymous {
						if st := ConvertInterfaceToGrpcStruct(field.Interface()).GetStructValue(); st != nil {
							for k, v := range st.GetFields() {
								fields[k] = v
							}
						}
					} else {
						fields[name] = ConvertInterfaceToGrpcStruct(field.Interface())
					}
				}
			}
		}
		return &structpb.Value{Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{Fields: fields}}}
	case reflect.Map:
		keyType := dataType.Key()
		isInt := assignableToInt(keyType)
		fields := make(map[string]*structpb.Value, val.Len())
		for _, key := range val.MapKeys() {
			val := val.MapIndex(key)
			var k string
			if isInt {
				k = strconv.Itoa(int(key.Int()))
			} else {
				k = key.String()
			}
			fields[k] = ConvertInterfaceToGrpcStruct(val.Interface())
		}
		return &structpb.Value{Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{Fields: fields}}}
	case reflect.Slice, reflect.Array:
		l := val.Len()
		list := &structpb.ListValue{}
		list.Values = make([]*structpb.Value, l)
		for i := 0; i < l; i++ {
			list.Values[i] = ConvertInterfaceToGrpcStruct(val.Index(i).Interface())
		}
		return &structpb.Value{Kind: &structpb.Value_ListValue{ListValue: list}}
	case reflect.String:
		return &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: val.String()}}
	case reflect.Bool:
		return &structpb.Value{Kind: &structpb.Value_BoolValue{BoolValue: val.Bool()}}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(val.Int())}}
	case reflect.Float32, reflect.Float64:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: val.Float()}}

	default:
		return nilValue
	}
}

func convertPrimitiveTypes(value interface{}) *structpb.Value {
	switch value.(type) {
	case bool:
		return &structpb.Value{Kind: &structpb.Value_BoolValue{BoolValue: value.(bool)}}
	case int, int16, int32, int64:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(reflect.ValueOf(value).Int())}}
	case float32, float64:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: reflect.ValueOf(value).Float()}}
	case string:
		return &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: value.(string)}}
	case types.Nil:
		return nilValue
	}
	return nil
}

func assignableToInt(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int:
		return true
	case reflect.Int8:
		return true
	case reflect.Int16:
		return true
	case reflect.Int32:
		return true
	case reflect.Int64:
		return true
	default:
		return false
	}
}

func getValue(data interface{}, pointer bool) (reflect.Value, bool) {
	val := reflect.ValueOf(data)
	if pointer {
		if val.IsNil() {
			return val, true
		} else {
			return val.Elem(), false
		}
	} else {
		return val, false
	}
}

func GetFieldName(fieldType reflect.StructField) (string, bool) {
	original := fieldType.Name
	transform := true
	if value, ok := fieldType.Tag.Lookup("json"); ok {
		opts := strings.Split(value, ",")
		if len(opts) > 0 {
			if opts[0] == "-" {
				return "", false
			}
			name := opts[0]
			if len(name) > 0 {
				original = name
				transform = false
			} else {
				transform = true
			}
		}
	}
	if transform {
		arr := []rune(original)
		arr[0] = unicode.ToLower(arr[0])
		original = string(arr)
	}
	return original, true
}
