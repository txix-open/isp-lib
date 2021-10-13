package validator

import (
	"context"

	"github.com/go-playground/validator/v10"
)

type Adapter struct {
	v *validator.Validate
}

func New() Adapter {
	return Adapter{validator.New()}
}

func (sv Adapter) Validate(ctx context.Context, v interface{}) (bool, map[string]string) {
	err := sv.v.StructCtx(ctx, v)
	if err == nil {
		return true, nil
	}
	validationErrors := err.(validator.ValidationErrors)
	descs := make(map[string]string, len(validationErrors))
	for _, e := range validationErrors {
		descs[e.Field()] = e.Tag()
	}
	return false, descs
}
