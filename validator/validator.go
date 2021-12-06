package validator

import (
	"strings"

	"github.com/asaskevich/govalidator"
)

type Adapter struct {
}

func New() Adapter {
	return Adapter{}
}

type wrapper struct {
	V interface{}
}

func (a Adapter) Validate(v interface{}) (bool, map[string]string) {
	ok, err := govalidator.ValidateStruct(wrapper{v}) //hack
	if ok {
		return true, nil
	}

	result := make(map[string]string)
	err = a.collectDetails(err, result)
	if err != nil {
		panic(err)
	}
	return false, result
}

func (a Adapter) collectDetails(err error, result map[string]string) error {
	switch e := err.(type) {
	case govalidator.Error:
		errName := e.Name
		if len(e.Path) > 0 {
			errName = strings.Join(append(e.Path, e.Name), ".")
			errName = errName[2:] //remove V.
		}
		result[errName] = e.Err.Error()
	case govalidator.Errors:
		for _, err := range e.Errors() {
			err = a.collectDetails(err, result)
			if err != nil {
				return err
			}
		}
	default:
		return err
	}
	return nil
}
