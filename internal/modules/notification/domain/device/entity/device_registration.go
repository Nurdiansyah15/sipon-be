package entity

import (
	"strings"
	"time"

	"sipon-be/internal/modules/notification/domain/device/constant"
	"sipon-be/internal/shared/kernel"
)

type DeviceRegistration struct {
	ID            string
	UserID        string
	Platform      constant.Platform
	PushProvider  constant.PushProvider
	ProviderToken string
	DeviceID      *string
	DeviceName    *string
	DeviceModel   *string
	OSVersion     *string
	AppVersion    *string
	Timezone      *string
	Active        bool
	LastSeenAt    time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type DeviceRegistrationParams struct {
	ID            string
	UserID        string
	Platform      constant.Platform
	PushProvider  constant.PushProvider
	ProviderToken string
	DeviceID      *string
	DeviceName    *string
	DeviceModel   *string
	OSVersion     *string
	AppVersion    *string
	Timezone      *string
}

func NewDeviceRegistration(params DeviceRegistrationParams) (*DeviceRegistration, error) {
	if strings.TrimSpace(params.ID) == "" {
		return nil, kernel.New(constant.CodeDeviceIDRequired)
	}
	if strings.TrimSpace(params.UserID) == "" {
		return nil, kernel.New(constant.CodeDeviceUserIDRequired)
	}
	if strings.TrimSpace(params.ProviderToken) == "" {
		return nil, kernel.New(constant.CodeDeviceTokenRequired)
	}
	if !params.Platform.IsValid() {
		return nil, kernel.New(constant.CodeDeviceInvalidPlatform)
	}
	if !params.PushProvider.IsValid() {
		return nil, kernel.New(constant.CodeDeviceInvalidProvider)
	}

	now := time.Now()
	return &DeviceRegistration{
		ID:            strings.TrimSpace(params.ID),
		UserID:        strings.TrimSpace(params.UserID),
		Platform:      params.Platform,
		PushProvider:  params.PushProvider,
		ProviderToken: strings.TrimSpace(params.ProviderToken),
		DeviceID:      normalizePtr(params.DeviceID),
		DeviceName:    normalizePtr(params.DeviceName),
		DeviceModel:   normalizePtr(params.DeviceModel),
		OSVersion:     normalizePtr(params.OSVersion),
		AppVersion:    normalizePtr(params.AppVersion),
		Timezone:      normalizePtr(params.Timezone),
		Active:        true,
		LastSeenAt:    now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func (dr *DeviceRegistration) UpdateToken(newToken string) {
	dr.ProviderToken = strings.TrimSpace(newToken)
	now := time.Now()
	dr.LastSeenAt = now
	dr.UpdatedAt = now
}

func (dr *DeviceRegistration) Deactivate() {
	dr.Active = false
	dr.UpdatedAt = time.Now()
}

func (dr *DeviceRegistration) Activate() {
	dr.Active = true
	dr.UpdatedAt = time.Now()
}

func (dr *DeviceRegistration) RecordSeen() {
	now := time.Now()
	dr.LastSeenAt = now
	dr.UpdatedAt = now
}

func normalizePtr(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
