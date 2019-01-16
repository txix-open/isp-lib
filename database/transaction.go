package database

import (
	"github.com/go-pg/pg"
	"github.com/integration-system/isp-lib/logger"
	"reflect"
)

func RunInTransaction(f interface{}) error {
	return RunInTransactionV2(dbManagerInstance.Db, f)
}

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
				err := "Invalid param type in callback. Expected struct with DB orm.DB field"
				logger.Error(err)
				panic(err)
			}
			field, present := param.FieldByName("DB")
			if !present || field.Type.String() != ormDbTypeName {
				err := "Invalid param type in callback. Expected struct with DB orm.DB field"
				logger.Error(err)
				panic(err)
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
		err := "Expected a function"
		logger.Error(err)
		panic(err)
	}
}
