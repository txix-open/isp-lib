package validate

import (
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CreateUnknownError(err error) error {
	logger.Error(err)
	st := status.New(codes.Unknown, utils.ServiceError)
	return st.Err()
}
