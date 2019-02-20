package http

import (
	"errors"
	"fmt"
	"reflect"
)

const (
	RestMType = MType("rest")
	SoapMType = MType("soap")
)

var (
	ctxType = reflect.TypeOf((*Ctx)(nil))

	ErrNotFunc     = errors.New("'handler' is not a function")
	ErrInvalidFunc = errors.New(`'handler' invalid. Expecting function with (
		[headers: map[string][string]],
		[requestBody: Any], 
		[ctx: *fasthttp.RequestContext]
	) params`)
)

type MType string

type funcDesc struct {
	inType     reflect.Type
	f          reflect.Value
	bodyNum    int
	headersNum int
	ctxNum     int
	inCount    int

	mType  MType
	method string
	uri    string
}

func (fd funcDesc) String() string {
	return fd.f.Type().String()
}

func toDesc(handler interface{}) (*funcDesc, error) {
	rt, rv := reflect.TypeOf(handler), reflect.ValueOf(handler)
	if rt.Kind() != reflect.Func {
		return nil, ErrNotFunc
	}
	fd := &funcDesc{f: rv, headersNum: -1, bodyNum: -1, ctxNum: -1}
	for i := 0; i < rt.NumIn(); i++ {
		inParam := rt.In(i)
		/*if inParam.Kind() == reflect.Map && fd.headersNum == -1 {
			fd.headersNum = i
			fd.inCount++
		} else*/if inParam == ctxType && fd.ctxNum == -1 {
			fd.ctxNum = i
			fd.inCount++
		} else if fd.bodyNum == -1 {
			fd.bodyNum = i
			fd.inType = inParam
			fd.inCount++
		} else {
			return nil, ErrInvalidFunc
		}
	}
	return fd, nil
}

func registerControllers(register func(uri, action string, mType MType, handler interface{}) error, uri string, controllers []interface{}) error {
	for i, h := range controllers {
		rv := reflect.ValueOf(h)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		rt := rv.Type()
		if rt.Kind() != reflect.Struct {
			return fmt.Errorf("'handlers[%d]' is not a struct", i)
		}
		for i := 0; i < rv.NumField(); i++ {
			val := rv.Field(i)
			fType := val.Type()
			if fType.Kind() == reflect.Func {
				fieldName := rt.Field(i).Name
				tag := rt.Field(i).Tag
				method, _ := tag.Lookup("method")
				t, _ := tag.Lookup("type")
				mType := MType(t)
				if mType != "" && mType != RestMType && mType != SoapMType {
					return fmt.Errorf("Unknown type: %s", mType)
				} else {
					mType = SoapMType
				}
				if val.IsNil() {
					return fmt.Errorf("Nil field: %s", fieldName)
				}
				if err := register(uri, method, mType, val.Interface()); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
