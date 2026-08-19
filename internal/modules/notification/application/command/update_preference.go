package command

import (
	"context"
	"time"

	"sipon-be/internal/modules/notification/application"
	"sipon-be/internal/modules/notification/application/dto"
	"sipon-be/internal/modules/notification/domain/preference/entity"
	prefRepo "sipon-be/internal/modules/notification/domain/preference/repository"
notifvo "sipon-be/internal/modules/notification/domain/valueobject"
	"sipon-be/internal/shared/kernel"
)

// UpdatePreferenceUseCase memperbarui preferensi notifikasi user.
type UpdatePreferenceUseCase struct {
	repo prefRepo.NotificationPreferenceRepository
}

func NewUpdatePreferenceUseCase(repo prefRepo.NotificationPreferenceRepository) *UpdatePreferenceUseCase {
	return &UpdatePreferenceUseCase{repo: repo}
}

func (uc *UpdatePreferenceUseCase) Execute(ctx context.Context, userID string, req dto.UpdatePreferenceRequest) (*dto.NotificationPreferenceResponse, error) {
	pref, err := uc.repo.FindOrCreateByUserID(ctx, userID)
	if err != nil {
		return nil, application.WrapDomainErr(err, "")
	}

	if req.AllNotificationsEnabled != nil {
		pref.AllEnabled = *req.AllNotificationsEnabled
	}

	if req.DoNotDisturbEnabled != nil {
		if *req.DoNotDisturbEnabled {
			startTime := req.DoNotDisturbStartTime
			endTime := req.DoNotDisturbEndTime
			if startTime == nil {
				startTime = pref.DNDStartTime
			}
			if endTime == nil {
				endTime = pref.DNDEndTime
			}
			if startTime == nil || endTime == nil {
				return nil, kernel.New(application.ErrCodeBadRequest)
			}
			if _, valErr := notifvo.NewDoNotDisturbWindow(*startTime, *endTime); valErr != nil {
				return nil, kernel.New(application.ErrCodeUnprocessable)
			}
			pref.DoNotDisturbEnabled = true
			pref.DNDStartTime = startTime
			pref.DNDEndTime = endTime
		} else {
			pref.DoNotDisturbEnabled = false
			pref.DNDStartTime = nil
			pref.DNDEndTime = nil
		}
	} else {
		if req.DoNotDisturbStartTime != nil || req.DoNotDisturbEndTime != nil {
			startTime := req.DoNotDisturbStartTime
			endTime := req.DoNotDisturbEndTime
			if startTime == nil {
				startTime = pref.DNDStartTime
			}
			if endTime == nil {
				endTime = pref.DNDEndTime
			}
			if startTime != nil && endTime != nil {
				if _, valErr := notifvo.NewDoNotDisturbWindow(*startTime, *endTime); valErr != nil {
					return nil, kernel.New(application.ErrCodeUnprocessable)
				}
			}
			pref.DNDStartTime = startTime
			pref.DNDEndTime = endTime
		}
	}

	pref.UpdatedAt = time.Now()

	if err := uc.repo.Update(ctx, pref); err != nil {
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
