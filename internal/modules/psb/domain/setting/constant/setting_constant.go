package constant

import "sipon-be/internal/shared/kernel"

type SettingStatus string

const (
	SettingStatusActive SettingStatus = "active"
	SettingStatusClosed SettingStatus = "closed"
)

const (
	ErrCodeInvalidSetting kernel.Code = "INVALID_SETTING"
)
