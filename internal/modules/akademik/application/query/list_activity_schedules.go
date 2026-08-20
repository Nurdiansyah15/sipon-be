package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	actEntity "sipon-be/internal/modules/akademik/domain/activity/entity"
	actRepo "sipon-be/internal/modules/akademik/domain/activity/repository"
	apEntity "sipon-be/internal/modules/akademik/domain/activity_period/entity"
	apRepo "sipon-be/internal/modules/akademik/domain/activity_period/repository"
	schRepo "sipon-be/internal/modules/akademik/domain/activity_schedule/repository"
	"sipon-be/internal/shared/kernel"
)

type ListActivitySchedulesUseCase struct {
	scheduleRepo       schRepo.ActivityScheduleRepository
	activityPeriodRepo apRepo.ActivityPeriodRepository
	activityRepo       actRepo.ActivityRepository
}

func NewListActivitySchedulesUseCase(
	scheduleRepo schRepo.ActivityScheduleRepository,
	activityPeriodRepo apRepo.ActivityPeriodRepository,
	activityRepo actRepo.ActivityRepository,
) *ListActivitySchedulesUseCase {
	return &ListActivitySchedulesUseCase{scheduleRepo: scheduleRepo, activityPeriodRepo: activityPeriodRepo, activityRepo: activityRepo}
}

func (uc *ListActivitySchedulesUseCase) Execute(ctx context.Context, activityPeriodID string) ([]dto.ActivityScheduleResponse, error) {
	schedules, err := uc.scheduleRepo.ListByActivityPeriod(ctx, activityPeriodID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	apIDs := make([]string, 0, len(schedules))
	for _, s := range schedules {
		apIDs = append(apIDs, s.ActivityPeriodID)
	}
	activityMap := map[string]*actEntity.Activity{}
	if aps, err := uc.activityPeriodRepo.FindByIDs(ctx, apIDs); err == nil {
		activityIDs := make([]string, 0, len(aps))
		for _, ap := range aps {
			activityIDs = append(activityIDs, ap.ActivityID)
		}
		if acts, err := uc.activityRepo.FindByIDs(ctx, activityIDs); err == nil {
			for _, a := range acts {
				activityMap[a.ID] = a
			}
		}
	}
	apMap := map[string]*apEntity.ActivityPeriod{}
	if aps, err := uc.activityPeriodRepo.FindByIDs(ctx, apIDs); err == nil {
		for _, ap := range aps {
			apMap[ap.ID] = ap
		}
	}

	items := make([]dto.ActivityScheduleResponse, len(schedules))
	for i, s := range schedules {
		resp := command.MapScheduleToDetailResponse(s)
		item := dto.ActivityScheduleResponse{
			ID:               resp.ID,
			ActivityPeriodID: resp.ActivityPeriodID,
			Type:             resp.Type,
			StartDate:        resp.StartDate,
			EndDate:          resp.EndDate,
			StartTime:        resp.StartTime,
			EndTime:          resp.EndTime,
			EarlyMinutes:     resp.EarlyMinutes,
			LateMinutes:      resp.LateMinutes,
			ReminderEarlyMinutes: resp.ReminderEarlyMinutes,
			CreatedAt:        resp.CreatedAt,
			UpdatedAt:        resp.UpdatedAt,
		}
		if ap, ok := apMap[s.ActivityPeriodID]; ok {
			if a, ok2 := activityMap[ap.ActivityID]; ok2 {
				item.ActivityName = a.Name
				item.ActivityCode = a.Code
			}
		}
		items[i] = item
	}
	return items, nil
}
