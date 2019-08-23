package structure

type GrpcError struct {
	ErrorMessage string        `json:"errorMessage"`
	ErrorCode    string        `json:"errorCode"`
	Details      []interface{} `json:"details"`
}
