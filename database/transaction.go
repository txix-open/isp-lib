package database

import (
	"errors"
	"reflect"

	"github.com/go-pg/pg/v9"
)

const (
	ormDbTypeName = "orm.DB"
)

func RunInTransactionV2(pdb *pg.DB, f interface{}) error {
	val := reflect.ValueOf(f)
	if val.Kind() == reflect.Func {
		t := val.Type()
		inParamCount := t.NumIn()
		params := make([]reflect.Value, inParamCount)
		tx, err := pdb.Begin()
		if err != nil {
			return err
		}
		for i := 0; i < inParamCount; i++ {
			param := t.In(i)
			if param.Kind() != reflect.Struct {
				panic(errors.New("invalid param type in callback. Expected struct with DB orm.DB field"))
			}
			field, present := param.FieldByName("DB")
			if !present || field.Type.String() != ormDbTypeName {
				panic(errors.New("invalid param type in callback. Expected struct with DB orm.DB field"))
			}
			repository := reflect.New(param)
			repository.Elem().FieldByName("DB").Set(reflect.ValueOf(tx))
			params[i] = repository.Elem()
		}
		var res []reflect.Value
		err = tx.RunInTransaction(func(tx *pg.Tx) error {
			res = val.Call(params)
			if len(res) == 0 {
				return nil
			} else {
				e := res[0].Interface()
				if err, ok := e.(error); ok {
					return err
				} else {
					return nil
				}
			}
		})
		return err
	} else {
		panic(errors.New("expected a function"))
	}
}
