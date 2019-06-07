package database

import (
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"reflect"
)

type CrudRepository struct {
	DB           orm.DB
	EntityType   reflect.Type
	CastToSlice  func(interface{}) interface{}
	CastToEntity func(interface{}) interface{}
}

func (r *CrudRepository) GetAllEntities(identities []int32) (interface{}, error) {
	entities := reflect.MakeSlice(reflect.SliceOf(r.EntityType), 0, 0).Interface()
	var err error
	result := r.CastToSlice(entities)
	query := r.DB.Model(result)
	if len(identities) > 0 {
		query = query.Where("id IN (?)", pg.In(identities))
	}
	err = query.Order("created_at DESC").
		Limit(500).
		Select()
	return result, err
}

func (r *CrudRepository) GetAllLongEntities(identities []int64) (interface{}, error) {
	entities := reflect.MakeSlice(reflect.SliceOf(r.EntityType), 0, 0).Interface()
	var err error
	result := r.CastToSlice(entities)
	query := r.DB.Model(result)
	if len(identities) > 0 {
		query = query.Where("id IN (?)", pg.In(identities))
	}
	err = query.Order("created_at DESC").
		Limit(500).
		Select()
	return result, err
}

func (r *CrudRepository) CreateEntity(entity interface{}) (interface{}, error) {
	castedEntity := r.CastToEntity(entity)
	_, err := r.DB.Model(castedEntity).
		Returning("*").
		Insert()
	return castedEntity, err
}

func (r *CrudRepository) UpdateEntity(entity interface{}) (interface{}, error) {
	castedEntity := r.CastToEntity(entity)
	_, err := r.DB.Model(castedEntity).WherePK().
		Returning("*").
		Update()
	return castedEntity, err
}

func (r *CrudRepository) DeleteEntities(identities []int32) (int, error) {
	castedEntity := r.CastToEntity(reflect.New(r.EntityType).Interface())
	result, err := r.DB.Model(castedEntity).
		Where("id IN (?)", pg.In(identities)).
		Delete()
	return result.RowsAffected(), err
}

func (r *CrudRepository) DeleteLongEntities(identities []int64) (int, error) {
	castedEntity := r.CastToEntity(reflect.New(r.EntityType).Interface())
	result, err := r.DB.Model(castedEntity).
		Where("id IN (?)", pg.In(identities)).
		Delete()
	return result.RowsAffected(), err
}

func (r *CrudRepository) GetEntityById(identity int32) (interface{}, error) {
	castedEntity := r.CastToEntity(reflect.New(r.EntityType).Interface())
	err := r.DB.Model(castedEntity).
		Where("id = ?", identity).
		First()
	if err != nil && err == pg.ErrNoRows {
		return castedEntity, nil
	}
	return castedEntity, err
}

func (r *CrudRepository) GetEntityByName(name string) (interface{}, error) {
	castedEntity := r.CastToEntity(reflect.New(r.EntityType).Interface())
	err := r.DB.Model(castedEntity).
		Where("name = ?", name).
		First()
	if err != nil && err == pg.ErrNoRows {
		return castedEntity, nil
	}
	return castedEntity, err
}
