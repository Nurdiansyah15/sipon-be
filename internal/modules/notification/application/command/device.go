package command

import (
	"context"
	"time"

	"sipon-be/internal/modules/notification/application"
	"sipon-be/internal/modules/notification/domain/device/constant"
	"sipon-be/internal/modules/notification/domain/device/entity"
	deviceRepo "sipon-be/internal/modules/notification/domain/device/repository"
	"github.com/google/uuid"
)

type RegisterDeviceUseCase struct {
	repo deviceRepo.DeviceRegistrationRepository
}

func NewRegisterDeviceUseCase(repo deviceRepo.DeviceRegistrationRepository) *RegisterDeviceUseCase {
	return &RegisterDeviceUseCase{repo: repo}
}

type RegisterDeviceInput struct {
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

func (uc *RegisterDeviceUseCase) Execute(ctx context.Context, input RegisterDeviceInput) (*entity.DeviceRegistration, error) {
	existing, err := uc.repo.FindByUserIDAndToken(ctx, input.UserID, input.ProviderToken)
	if err == nil && existing != nil {
		existing.RecordSeen()
		if input.Timezone != nil {
			existing.Timezone = input.Timezone
		}
		if err := uc.repo.Update(ctx, existing); err != nil {
			return nil, application.WrapDomainErr(err, constant.CodeDevicePersistenceFailed)
		}
		return existing, nil
	}

	dr, err := entity.NewDeviceRegistration(entity.DeviceRegistrationParams{
		ID:            uuid.NewString(),
		UserID:        input.UserID,
		Platform:      input.Platform,
		PushProvider:  input.PushProvider,
		ProviderToken: input.ProviderToken,
		DeviceID:      input.DeviceID,
		DeviceName:    input.DeviceName,
		DeviceModel:   input.DeviceModel,
		OSVersion:     input.OSVersion,
		AppVersion:    input.AppVersion,
		Timezone:      input.Timezone,
	})
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Save(ctx, dr); err != nil {
		return nil, application.WrapDomainErr(err, constant.CodeDevicePersistenceFailed)
	}

	return dr, nil
}

type UnregisterDeviceUseCase struct {
	repo deviceRepo.DeviceRegistrationRepository
}

func NewUnregisterDeviceUseCase(repo deviceRepo.DeviceRegistrationRepository) *UnregisterDeviceUseCase {
	return &UnregisterDeviceUseCase{repo: repo}
}

func (uc *UnregisterDeviceUseCase) Execute(ctx context.Context, userID, token string) error {
	dr, err := uc.repo.FindByUserIDAndToken(ctx, userID, token)
	if err != nil {
		return application.WrapDomainErr(err, constant.CodeDeviceNotFound)
	}
	dr.Deactivate()
	if err := uc.repo.Update(ctx, dr); err != nil {
		return application.WrapDomainErr(err, constant.CodeDevicePersistenceFailed)
	}
	return nil
}

type ListDevicesUseCase struct {
	repo deviceRepo.DeviceRegistrationRepository
}

func NewListDevicesUseCase(repo deviceRepo.DeviceRegistrationRepository) *ListDevicesUseCase {
	return &ListDevicesUseCase{repo: repo}
}

func (uc *ListDevicesUseCase) Execute(ctx context.Context, userID string) ([]*entity.DeviceRegistration, error) {
	return uc.repo.FindByUserID(ctx, userID, false)
}

type DeactivateInvalidTokensUseCase struct {
	repo deviceRepo.DeviceRegistrationRepository
}

func NewDeactivateInvalidTokensUseCase(repo deviceRepo.DeviceRegistrationRepository) *DeactivateInvalidTokensUseCase {
	return &DeactivateInvalidTokensUseCase{repo: repo}
}

func (uc *DeactivateInvalidTokensUseCase) Execute(ctx context.Context, tokens []string) {
	for _, token := range tokens {
		_ = uc.repo.DeactivateByToken(ctx, token)
	}
}

type CleanupStaleDevicesUseCase struct {
	repo deviceRepo.DeviceRegistrationRepository
}

func NewCleanupStaleDevicesUseCase(repo deviceRepo.DeviceRegistrationRepository) *CleanupStaleDevicesUseCase {
	return &CleanupStaleDevicesUseCase{repo: repo}
}

func (uc *CleanupStaleDevicesUseCase) Execute(ctx context.Context, staleDuration time.Duration) {
	// TODO: implement cleanup of devices not seen in staleDuration
	_ = staleDuration
}
