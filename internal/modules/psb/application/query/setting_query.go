package query

import (
	"context"

	"sipon-be/internal/modules/psb/application"
	"sipon-be/internal/modules/psb/application/dto"
	sentity "sipon-be/internal/modules/psb/domain/setting/entity"
	srepo 	"sipon-be/internal/modules/psb/domain/setting/repository"
	"sipon-be/internal/shared/kernel"
)

type SettingQueryUseCase struct {
	settingRepo srepo.PsbSettingRepository
}

func NewSettingQueryUseCase(settingRepo srepo.PsbSettingRepository) *SettingQueryUseCase {
	return &SettingQueryUseCase{settingRepo: settingRepo}
}

func (uc *SettingQueryUseCase) GetActive(ctx context.Context) (*dto.SettingResponse, error) {
	s, err := uc.settingRepo.FindActive(ctx)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}
	return mapSettingToResponse(s), nil
}

func (uc *SettingQueryUseCase) GetByID(ctx context.Context, id string) (*dto.SettingResponse, error) {
	s, err := uc.settingRepo.FindByID(ctx, id)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}
	return mapSettingToResponse(s), nil
}

func (uc *SettingQueryUseCase) List(ctx context.Context) ([]dto.SettingResponse, error) {
	settings, err := uc.settingRepo.List(ctx)
	if err != nil {
		return nil, nil
	}

	items := make([]dto.SettingResponse, len(settings))
	for i, s := range settings {
		items[i] = *mapSettingToResponse(s)
	}

	return items, nil
}

func mapSettingToResponse(s *sentity.PsbSetting) *dto.SettingResponse {
	return &dto.SettingResponse{
		ID:           s.ID,
		Name:         s.Name,
		StartPeriod:  s.StartPeriod.Format("2006-01-02"),
		EndPeriod:    s.EndPeriod.Format("2006-01-02"),
		Status:       string(s.Status),
		Quota:        s.Quota,
		RegFee:       s.RegFee,
		BankAccounts: s.BankAccounts,
		DataPurgedAt: s.DataPurgedAt,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}
}
