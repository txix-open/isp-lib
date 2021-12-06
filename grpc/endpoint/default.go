package endpoint

import (
	"github.com/integration-system/isp-lib/v3/log"
	"github.com/integration-system/isp-lib/v3/validator"
)

func Default(logger log.Logger) Mapper {
	paramMappers := []ParamMapper{
		ContextParam(),
		AuthDataParam(),
	}
	return NewMapper(
		paramMappers,
		JsonRequestExtractor{validator: validator.Default},
		JsonResponseMapper{},
	).WithMiddlewares(
		RequestId(),
		ErrorHandler(logger),
		Recovery(),
	)
}
