package structure

import (
	"github.com/integration-system/isp-lib/utils"
	"google.golang.org/grpc/metadata"
)

type Isolation metadata.MD

func (i Isolation) GetInstanceId() (string, error) {
	return utils.ResolveMetadataIdentityV2(utils.InstanceIdHeader, metadata.MD(i))
}

func (i Isolation) GetSystemId() (int32, error) {
	val, err := utils.ResolveMetadataIdentity(utils.SystemIdHeader, metadata.MD(i))
	return int32(val), err
}

func (i Isolation) GetDomainId() (int32, error) {
	val, err := utils.ResolveMetadataIdentity(utils.DomainIdHeader, metadata.MD(i))
	return int32(val), err
}

func (i Isolation) GetServiceId() (int32, error) {
	val, err := utils.ResolveMetadataIdentity(utils.ServiceIdHeader, metadata.MD(i))
	return int32(val), err
}

func (i Isolation) GetApplicationId() (int32, error) {
	val, err := utils.ResolveMetadataIdentity(utils.ApplicationIdHeader, metadata.MD(i))
	return int32(val), err
}

func (i Isolation) GetUserId() (int64, error) {
	val, err := utils.ResolveMetadataIdentity(utils.UserIdHeader, metadata.MD(i))
	return int64(val), err
}

func (i Isolation) GetDeviceId() (int64, error) {
	val, err := utils.ResolveMetadataIdentity(utils.DeviceIdHeader, metadata.MD(i))
	return int64(val), err
}

func (i Isolation) GetUserToken() (string, error) {
	return utils.ResolveMetadataIdentityV2(utils.UserTokenHeaderLC, metadata.MD(i))
}
