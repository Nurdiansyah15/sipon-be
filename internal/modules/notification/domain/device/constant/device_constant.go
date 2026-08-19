package constant

import "sipon-be/internal/shared/kernel"

type Platform string

const (
	PlatformAndroid Platform = "android"
	PlatformIOS     Platform = "ios"
	PlatformWeb     Platform = "web"
)

func (p Platform) IsValid() bool {
	switch p {
	case PlatformAndroid, PlatformIOS, PlatformWeb:
		return true
	}
	return false
}

type PushProvider string

const (
	PushProviderFCM     PushProvider = "fcm"
	PushProviderAPNS    PushProvider = "apns"
	PushProviderHuawei  PushProvider = "huawei"
	PushProviderWebPush PushProvider = "web_push"
)

func (p PushProvider) IsValid() bool {
	switch p {
	case PushProviderFCM, PushProviderAPNS, PushProviderHuawei, PushProviderWebPush:
		return true
	}
	return false
}

const (
	CodeDeviceIDRequired        kernel.Code = "DEVICE_ID_REQUIRED"
	CodeDeviceUserIDRequired    kernel.Code = "DEVICE_USER_ID_REQUIRED"
	CodeDeviceTokenRequired     kernel.Code = "DEVICE_TOKEN_REQUIRED"
	CodeDeviceInvalidPlatform   kernel.Code = "DEVICE_INVALID_PLATFORM"
	CodeDeviceInvalidProvider   kernel.Code = "DEVICE_INVALID_PROVIDER"
	CodeDeviceNotFound          kernel.Code = "DEVICE_NOT_FOUND"
	CodeDeviceDuplicateToken    kernel.Code = "DEVICE_DUPLICATE_TOKEN"
	CodeDevicePersistenceFailed kernel.Code = "DEVICE_PERSISTENCE_FAILED"
)
