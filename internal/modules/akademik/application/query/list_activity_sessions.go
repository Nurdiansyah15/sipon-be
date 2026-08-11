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
	schEntity "sipon-be/internal/modules/akademik/domain/activity_schedule/entity"
	schRepo "sipon-be/internal/modules/akademik/domain/activity_schedule/repository"
	sesEntity "sipon-be/internal/modules/akademik/domain/activity_session/entity"
	sesRepo "sipon-be/internal/modules/akademik/domain/activity_session/repository"
	"sipon-be/internal/shared/kernel"
)

type ListActivitySessionsUseCase struct {
	sessionRepo        sesRepo.ActivitySessionRepository
	scheduleRepo       schRepo.ActivityScheduleRepository
	activityPeriodRepo apRepo.ActivityPeriodRepository
	activityRepo       actRepo.ActivityRepository
}

func NewListActivitySessionsUseCase(
	sessionRepo sesRepo.ActivitySessionRepository,
	scheduleRepo schRepo.ActivityScheduleRepository,
	activityPeriodRepo apRepo.ActivityPeriodRepository,
	activityRepo actRepo.ActivityRepository,
) *ListActivitySessionsUseCase {
	return &ListActivitySessionsUseCase{sessionRepo: sessionRepo, scheduleRepo: scheduleRepo, activityPeriodRepo: activityPeriodRepo, activityRepo: activityRepo}
}

func (uc *ListActivitySessionsUseCase) Execute(ctx context.Context, q dto.ActivitySessionListQuery) ([]dto.ActivitySessionResponse, *dto.Meta, error) {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.Limit == 0 {
		q.Limit = 20
	}

	result, err := uc.sessionRepo.List(ctx, sesRepo.ActivitySessionListQuery{
		ActivityScheduleID: q.ActivityScheduleID,
		AcademicPeriodID:   q.AcademicPeriodID,
		Status:             q.Status,
		StartDate:          q.StartDate,
		EndDate:            q.EndDate,
		Page:               q.Page,
		Limit:              q.Limit,
	})
	if err != nil {
		return nil, nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	scheduleMap, apMap, activityMap := uc.enrich(ctx, result.Items)

	items := make([]dto.ActivitySessionResponse, len(result.Items))
	for i, s := range result.Items {
		resp := command.MapSessionToResponse(s)
		if sch, ok := scheduleMap[s.ActivityScheduleID]; ok {
			resp.ScheduleType = string(sch.Type)
			if ap, ok2 := apMap[sch.ActivityPeriodID]; ok2 {
				if a, ok3 := activityMap[ap.ActivityID]; ok3 {
					resp.ActivityName = a.Name
					resp.ActivityCode = a.Code
				}
			}
		}
		items[i] = *resp
	}
	return items, dto.NewMeta(q.Page, q.Limit, result.Total), nil
}

func (uc *ListActivitySessionsUseCase) enrich(ctx context.Context, items []*sesEntity.ActivitySession) (map[string]*schEntity.ActivitySchedule, map[string]*apEntity.ActivityPeriod, map[string]*actEntity.Activity) {
	scheduleIDs := map[string]struct{}{}
	for _, s := range items {
		scheduleIDs[s.ActivityScheduleID] = struct{}{}
	}
	schedules, _ := uc.scheduleRepo.FindByIDs(ctx, idsFromSet(scheduleIDs))
	scheduleMap := make(map[string]*schEntity.ActivitySchedule, len(schedules))
	apIDs := map[string]struct{}{}
	for _, sch := range schedules {
		scheduleMap[sch.ID] = sch
		apIDs[sch.ActivityPeriodID] = struct{}{}
	}

	aps, _ := uc.activityPeriodRepo.FindByIDs(ctx, idsFromSet(apIDs))
	apMap := make(map[string]*apEntity.ActivityPeriod, len(aps))
	activityIDs := map[string]struct{}{}
	for _, ap := range aps {
		apMap[ap.ID] = ap
		activityIDs[ap.ActivityID] = struct{}{}
	}

	acts, _ := uc.activityRepo.FindByIDs(ctx, idsFromSet(activityIDs))
	activityMap := make(map[string]*actEntity.Activity, len(acts))
	for _, a := range acts {
		activityMap[a.ID] = a
	}
	return scheduleMap, apMap, activityMap
}
