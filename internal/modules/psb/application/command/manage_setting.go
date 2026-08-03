package command

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/psb/application"
	"sipon-be/internal/modules/psb/application/dto"
	sentity "sipon-be/internal/modules/psb/domain/setting/entity"
	srepo "sipon-be/internal/modules/psb/domain/setting/repository"
	"sipon-be/internal/shared/kernel"
)

type ManageSettingUseCase struct {
	settingRepo srepo.PsbSettingRepository
}

func NewManageSettingUseCase(settingRepo srepo.PsbSettingRepository) *ManageSettingUseCase {
	return &ManageSettingUseCase{settingRepo: settingRepo}
}

func (uc *ManageSettingUseCase) Create(ctx context.Context, req dto.CreateSettingRequest) (*dto.SettingResponse, error) {
	startPeriod, err := time.Parse("2006-01-02", req.StartPeriod)
	if err != nil {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}
	endPeriod, err := time.Parse("2006-01-02", req.EndPeriod)
	if err != nil {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}

	quota := json.RawMessage("{}")
	if req.Quota != nil {
		quota = req.Quota
	}
	bankAccounts := json.RawMessage("[]")
	if req.BankAccounts != nil {
		bankAccounts = req.BankAccounts
	}

	s, err := sentity.NewPsbSetting(uuid.NewString(), req.Name, startPeriod, endPeriod, quota, bankAccounts, req.RegFee)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.settingRepo.Save(ctx, s); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return mapSettingToResponse(s), nil
}

func (uc *ManageSettingUseCase) Update(ctx context.Context, id string, req dto.UpdateSettingRequest) (*dto.SettingResponse, error) {
	s, err := uc.settingRepo.FindByID(ctx, id)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	if req.Name != nil {
		s.Name = *req.Name
	}
	if req.Status != nil && *req.Status == "closed" {
		if err := s.Close(); err != nil {
			return nil, kernel.New(application.ErrCodeConflict)
		}
	}
	if req.RegFee != nil {
		s.RegFee = *req.RegFee
	}
	if req.Quota != nil {
		s.Quota = req.Quota
	}
	if req.BankAccounts != nil {
		s.BankAccounts = req.BankAccounts
	}
	s.UpdatedAt = time.Now()

	if err := uc.settingRepo.Update(ctx, s); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return mapSettingToResponse(s), nil
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
