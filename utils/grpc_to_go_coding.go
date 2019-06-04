package utils

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/ptypes/struct"
	"reflect"
	"strconv"
	"time"
)

type GrpcValueUnmarshaler interface {
	FromGrpcValue(*structpb.Value) (bool, error)
}

func ConvertGrpcStructToMap(data map[string]*structpb.Value) map[string]interface{} {
	fields := make(map[string]interface{}, len(data))
	for key, value := range data {
		fields[key] = ConvertGrpcStructToInterface(value)
	}
	return fields
}

func ConvertGrpcStructToInterface(data *structpb.Value) interface{} {
	if data == nil {
		return nil
	}
	switch data.GetKind().(type) {
	case *structpb.Value_StructValue:
		dataStruct := data.GetKind().(*structpb.Value_StructValue).StructValue.Fields
		fields := make(map[string]interface{}, len(dataStruct))
		for key, value := range dataStruct {
			fields[key] = ConvertGrpcStructToInterface(value)
		}
		return fields
	case *structpb.Value_ListValue:
		dataList := data.GetKind().(*structpb.Value_ListValue).ListValue.Values
		list := make([]interface{}, len(dataList))
		for i, value := range dataList {
			list[i] = ConvertGrpcStructToInterface(value)
		}
		return list
	case *structpb.Value_StringValue:
		return data.GetKind().(*structpb.Value_StringValue).StringValue
	case *structpb.Value_NumberValue:
		return data.GetKind().(*structpb.Value_NumberValue).NumberValue
	case *structpb.Value_BoolValue:
		return data.GetKind().(*structpb.Value_BoolValue).BoolValue
	case *structpb.Value_NullValue:
		return nil

	default:
		return nil
	}

}

func ConvertGrpcToGo(data *structpb.Value, ptr interface{}) error {
	rt, rv := reflect.TypeOf(ptr), reflect.ValueOf(ptr)
	kind := rt.Kind()
	if kind != reflect.Ptr || rv.IsNil() {
		return errors.New("Expected not nil pointer")
	}
	_, err := grpcToGoInner(data, rv)
	if err != nil {
		return err
	}
	return nil
}

func grpcToGoInner(data *structpb.Value, ptr reflect.Value) (bool, error) {
	if um, ok := ptr.Interface().(GrpcValueUnmarshaler); ok {
		return um.FromGrpcValue(data)
	}

	if isNil := getNullValue(data); isNil {
		return isNil, nil
	}

	val := ptr.Elem()
	t := val.Type()
	switch t.Kind() {
	case reflect.Struct:
		if _, ok := val.Interface().(time.Time); ok {
			if dataString := data.GetStringValue(); dataString != "" {
				t, err := time.Parse(FullDateFormat, dataString)
				if err != nil {
					return false, err
				}
				val.Set(reflect.ValueOf(t))
				return false, nil
			} else {
				return false, errors.New(fmt.Sprintf("Expected string in format %s. Got: %s", FullDateFormat, data.Kind))
			}
		}
		if dataStruct := data.GetStructValue(); dataStruct != nil {
			fieldsCount := t.NumField()
			for i := 0; i < fieldsCount; i++ {
				fieldDesc := t.Field(i)
				fieldType := fieldDesc.Type
				isPtr := fieldType.Kind() == reflect.Ptr
				if isPtr {
					fieldType = fieldType.Elem()
				}
				name, accept := GetFieldName(fieldDesc)
				if !accept {
					continue
				}
				dataValue, ok := dataStruct.Fields[name]
				if fieldDesc.Anonymous {
					dataValue = data
					ok = true
				}
				if ok {
					field := val.Field(i)
					fieldValue := field
					if isPtr {
						fieldValue = reflect.New(fieldType)
					} else {
						fieldValue = fieldValue.Addr()
					}
					if isNil, err := grpcToGoInner(dataValue, fieldValue); err != nil {
						return false, err
					} else if isNil {
						continue
					}
					if isPtr {
						field.Set(fieldValue)
					}
				}
			}
		} else {
			return false, errors.New(fmt.Sprintf("Expected a struct. Got: %s", data.Kind))
		}
	case reflect.Map:
		if dataStruct := data.GetStructValue(); dataStruct != nil {
			m := reflect.MakeMapWithSize(t, len(dataStruct.Fields))
			elemType := t.Elem()
			keyType := t.Key()
			isInt := assignableToInt(keyType)
			isPtr := elemType.Kind() == reflect.Ptr
			if isPtr {
				elemType = elemType.Elem()
			}
			for key, value := range dataStruct.Fields {
				elemValue := reflect.New(elemType)
				if isNil, err := grpcToGoInner(value, elemValue); err != nil {
					return false, err
				} else if isNil {
					continue
				}
				var k reflect.Value
				if isInt {
					if i, err := strconv.Atoi(key); err != nil {
						return false, err
					} else {
						k = reflect.ValueOf(int64(i))
					}
				} else {
					k = reflect.ValueOf(key)
				}
				if isPtr {
					m.SetMapIndex(k, elemValue)
				} else {
					m.SetMapIndex(k, elemValue.Elem())
				}
			}
			val.Set(m)
		} else {
			return false, errors.New(fmt.Sprintf("Expected a struct. Got: %s", data.Kind))
		}
	case reflect.Slice:
		if dataList := data.GetListValue(); dataList != nil {
			l := len(dataList.Values)
			slice := reflect.MakeSlice(t, l, l)
			elemType := t.Elem()
			isPtr := elemType.Kind() == reflect.Ptr
			if isPtr {
				elemType = elemType.Elem()
			}
			for i, value := range dataList.Values {
				elemValue := slice.Index(i)
				if isPtr {
					elemValue = reflect.New(elemType)
				} else {
					elemValue = elemValue.Addr()
				}
				if isNil, err := grpcToGoInner(value, elemValue); err != nil {
					return false, err
				} else if isNil {
					continue
				}
				if isPtr {
					slice.Index(i).Set(elemValue)
				}
			}
			val.Set(slice)
		} else {
			return false, errors.New(fmt.Sprintf("Expected a array. Got: %s", data.Kind))
		}
	case reflect.String:
		if str, err := getStringValue(data); err == nil {
			val.SetString(str)
		} else {
			return false, err
		}
	case reflect.Bool:
		if b, err := getBoolValue(data); err == nil {
			val.SetBool(b)
		} else {
			return false, err
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if num, err := getNumberValue(data); err == nil {
			val.SetInt(int64(num))
		} else {
			return false, err
		}
	case reflect.Float32, reflect.Float64:
		if num, err := getNumberValue(data); err == nil {
			val.SetFloat(num)
		} else {
			return false, err
		}
	case reflect.Interface:
		val.Set(reflect.ValueOf(ConvertGrpcStructToInterface(data)))
	default:
		return false, nil
	}

	return false, nil
}

func getNumberValue(data *structpb.Value) (float64, error) {
	if x, ok := data.GetKind().(*structpb.Value_NumberValue); ok {
		return x.NumberValue, nil
	}
	return 0, errors.New(fmt.Sprintf("Expected a number. Got: %s", data.Kind))
}

func getStringValue(data *structpb.Value) (string, error) {
	if x, ok := data.GetKind().(*structpb.Value_StringValue); ok {
		return x.StringValue, nil
	}
	return "", errors.New(fmt.Sprintf("Expected a string. Got: %s", data.Kind))
}

func getBoolValue(data *structpb.Value) (bool, error) {
	if x, ok := data.GetKind().(*structpb.Value_BoolValue); ok {
		return x.BoolValue, nil
	}
	return false, errors.New(fmt.Sprintf("Expected a boolean. Got: %s", data.Kind))
}

func getNullValue(data *structpb.Value) bool {
	_, ok := data.GetKind().(*structpb.Value_NullValue)
	return ok
}
