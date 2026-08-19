package query

import (
	"context"

	"sipon-be/internal/modules/notification/application"
	"sipon-be/internal/modules/notification/application/dto"
	"sipon-be/internal/modules/notification/domain/preference/entity"
	prefRepo "sipon-be/internal/modules/notification/domain/preference/repository"
)

type GetPreferenceUseCase struct {
	repo prefRepo.NotificationPreferenceRepository
}

func NewGetPreferenceUseCase(repo prefRepo.NotificationPreferenceRepository) *GetPreferenceUseCase {
	return &GetPreferenceUseCase{repo: repo}
}

func (uc *GetPreferenceUseCase) Execute(ctx context.Context, userID string) (*dto.NotificationPreferenceResponse, error) {
	pref, err := uc.repo.FindOrCreateByUserID(ctx, userID)
	if err != nil {
		return nil, application.WrapDomainErr(err, "")
	}
	return mapPrefToResponse(pref), nil
}

func mapPrefToResponse(p *entity.NotificationPreference) *dto.NotificationPreferenceResponse {
	return &dto.NotificationPreferenceResponse{
		ID:                      p.ID,
		UserID:                  p.UserID,
		AllNotificationsEnabled: p.AllEnabled,
		DoNotDisturbEnabled:     p.DoNotDisturbEnabled,
		DoNotDisturbStartTime:   p.DNDStartTime,
		DoNotDisturbEndTime:     p.DNDEndTime,
		CreatedAt:               dto.FormatRFC3339(p.CreatedAt),
		UpdatedAt:               dto.FormatRFC3339(p.UpdatedAt),
	}
}
