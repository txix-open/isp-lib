package http

import (
	"google.golang.org/grpc/codes"
	"net/http"
)

func init() {
	for httpCode, grpcCode := range codeMap {
		inverseCodeMap[grpcCode] = httpCode
	}
}

var (
	codeMap = map[int]codes.Code{
		http.StatusOK:                  codes.OK,
		http.StatusRequestTimeout:      codes.Canceled,
		http.StatusBadRequest:          codes.InvalidArgument,
		http.StatusGatewayTimeout:      codes.DeadlineExceeded,
		http.StatusNotFound:            codes.NotFound,
		http.StatusConflict:            codes.AlreadyExists,
		http.StatusForbidden:           codes.PermissionDenied,
		http.StatusUnauthorized:        codes.Unauthenticated,
		http.StatusTooManyRequests:     codes.ResourceExhausted,
		http.StatusPreconditionFailed:  codes.FailedPrecondition,
		http.StatusNotImplemented:      codes.Unimplemented,
		http.StatusInternalServerError: codes.Internal,
		http.StatusServiceUnavailable:  codes.Unavailable,
	}
	inverseCodeMap = map[codes.Code]int{}
)

func HttpStatusToCode(status int) codes.Code {
	code, ok := codeMap[status]
	if !ok {
		return codes.Unknown
	}
	return code
}

func CodeToHttpStatus(code codes.Code) int {
	status, ok := inverseCodeMap[code]
	if !ok {
		return http.StatusInternalServerError
	}
	return status
}
