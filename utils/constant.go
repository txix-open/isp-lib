package utils

import (
	"os"
	"strings"
)

var (
	// ===== ENV =====
	EnvProfileName   = os.Getenv("APP_PROFILE")
	EnvConfigPath    = os.Getenv("APP_CONFIG_PATH")
	EnvMigrationPath = os.Getenv("APP_MIGRATION_PATH")
	DEV              = strings.ToLower(os.Getenv("APP_MODE")) == "dev"
	LOG_LEVEL        = os.Getenv("LOG_LEVEL")

	FileLoggerConfig = os.Getenv("LOG_CONFIG")
)

const (
	ErrorHeader      = "X-MP-ERROR"
	ErrorsHeader     = "X-MP-ERRORS"
	ExpectFileHeader = "X-EXPECT-FILE"

	// ===== ERRORS =====
	ServiceError    = "Service is not available now, please try later"
	ValidationError = "Validation errors"
	Required        = "Required"

	// ===== DATE =====
	FullDateFormat = "2006-01-02T15:04:05.999-07:00"

	// ===== GRPC =====
	ProxyMethodNameHeader = "proxy_method_name"
	MethodDefaultGroup    = "api/"

	// ===== ADMIN =====
	ADMIN_AUTH_HEADER_NAME = "x-auth-admin"
	DB_SCHEME              = "admin_service"

	ApplicationIdHeader = "x-application-identity"
	UserIdHeader        = "x-user-identity"
	DeviceIdHeader      = "x-device-identity"
	ServiceIdHeader     = "x-service-identity"
	DomainIdHeader      = "x-domain-identity"
	SystemIdHeader      = "x-system-identity"
	InstanceIdHeader    = "x-instance-identity"

	ApplicationTokenHeader = "X-APPLICATION-TOKEN"
	UserTokenHeader        = "X-USER-TOKEN"
	DeviceTokenHeader      = "X-DEVICE-TOKEN"

	HeaderNotSpecifiedValue = "not_specified"
)

var (
	UserTokenHeaderLC = strings.ToLower(UserTokenHeader)
)
