package utils

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strconv"
)

func ResolveMetadataIdentity(key string, md metadata.MD) (int, error) {
	if arr, ok := md[key]; ok && len(arr) > 0 && arr[0] != HeaderNotSpecifiedValue {
		res, err := strconv.Atoi(arr[0])
		if err != nil {
			return 0, status.Errorf(codes.DataLoss, "Invalid metadata value [%s]. Expected integer", key)
		}
		return res, nil
	} else {
		return 0, status.Errorf(codes.DataLoss, "Metadata [%s] is required", key)
	}
}

func ResolveMetadataIdentityV2(key string, md metadata.MD) (string, error) {
	if arr, ok := md[key]; ok && len(arr) > 0 && arr[0] != HeaderNotSpecifiedValue {
		return arr[0], nil
	} else {
		return "", status.Errorf(codes.DataLoss, "Metadata [%s] is required", key)
	}
}
