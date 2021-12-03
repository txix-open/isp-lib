package endpoint

import (
	"github.com/integration-system/isp-lib/v3/log"
	"github.com/integration-system/isp-lib/v3/validator"
)

func DefaultMapper(logger log.Logger) Mapper {
	paramMappers := []ParamMapper{
		ContextParam(),
	}
	return NewMapper(
		paramMappers,
		JsonRequestExtractor{},
		JsonResponseMapper{},
	).WithMiddlewares(
		RequestId(),
		ErrorLogger(logger),
		Recovery(),
		RequestBodyValidationMiddleware(validator.Default),
	)
}
