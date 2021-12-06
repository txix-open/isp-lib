package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/asaskevich/govalidator"
)

type Adapter struct {
}

func New() Adapter {
	return Adapter{}
}

func (a Adapter) Validate(v interface{}) (bool, map[string]string) {
	rv := reflect.ValueOf(v)
	rt := rv.Type()
	if rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array {
		for i := 0; i < rv.Len(); i++ {
			item := rv.Index(i)
			ok, details := a.validate(item.Interface())
			if !ok {
				newDetails := make(map[string]string, len(details))
				for k, v := range details {
					newDetails[fmt.Sprintf("%d.%s", i, k)] = v
				}
				return false, newDetails
			}
		}
		return true, nil
	}

	return a.validate(v)
}

func (a Adapter) validate(v interface{}) (bool, map[string]string) {
	ok, err := govalidator.ValidateStruct(v)
	if ok {
		return true, nil
	}

	result := make(map[string]string)
	a.collectDetails(err, result)
	return false, result
}

func (a Adapter) collectDetails(err error, result map[string]string) {
	switch e := err.(type) {
	case govalidator.Error:
		errName := e.Name
		if len(e.Path) > 0 {
			errName = strings.Join(append(e.Path, e.Name), ".")
		}
		result[errName] = e.Err.Error()
	case govalidator.Errors:
		for _, err := range e.Errors() {
			a.collectDetails(err, result)
		}
	}
}
