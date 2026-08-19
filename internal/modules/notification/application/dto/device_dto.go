package dto

type RegisterDeviceRequest struct {
	Platform      string `json:"platform" binding:"required"`
	PushProvider  string `json:"push_provider" binding:"required"`
	ProviderToken string `json:"provider_token" binding:"required"`
	DeviceID      string `json:"device_id"`
	DeviceName    string `json:"device_name"`
	DeviceModel   string `json:"device_model"`
	OSVersion     string `json:"os_version"`
	AppVersion    string `json:"app_version"`
	Timezone      string `json:"timezone"`
}

type UnregisterDeviceRequest struct {
	ProviderToken string `json:"provider_token" binding:"required"`
}

type DeviceResponse struct {
	ID            string `json:"id"`
	Platform      string `json:"platform"`
	PushProvider  string `json:"push_provider"`
	DeviceName    string `json:"device_name,omitempty"`
	DeviceModel   string `json:"device_model,omitempty"`
	OSVersion     string `json:"os_version,omitempty"`
	AppVersion    string `json:"app_version,omitempty"`
	Active        bool   `json:"active"`
	LastSeenAt    string `json:"last_seen_at"`
}
