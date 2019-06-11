package http

import (
	"github.com/asaskevich/govalidator"
	"github.com/integration-system/isp-lib/utils"
	"net/http"
)

type ValidationErrors struct {
	*RESTFault
	Details map[string]string
}

func validate(ctx *Ctx, value interface{}) error {
	err := utils.ValidateV2(value)
	if err == nil {
		return nil
	}
	m := govalidator.ErrorsByField(err)
	return &ValidationErrors{
		RESTFault: &RESTFault{
			Code:   http.StatusBadRequest,
			Status: http.StatusText(http.StatusBadRequest),
		},
		Details: m,
	}
}
