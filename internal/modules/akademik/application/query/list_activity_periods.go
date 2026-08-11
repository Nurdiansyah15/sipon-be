package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	periodRepo "sipon-be/internal/modules/akademik/domain/academic_period/repository"
	actEntity "sipon-be/internal/modules/akademik/domain/activity/entity"
	actRepo "sipon-be/internal/modules/akademik/domain/activity/repository"
	apEntity "sipon-be/internal/modules/akademik/domain/activity_period/entity"
	apRepo "sipon-be/internal/modules/akademik/domain/activity_period/repository"
	"sipon-be/internal/shared/kernel"
)

type ListActivityPeriodsUseCase struct {
	activityPeriodRepo apRepo.ActivityPeriodRepository
	activityRepo       actRepo.ActivityRepository
	periodRepo         periodRepo.AcademicPeriodRepository
}

func NewListActivityPeriodsUseCase(
	activityPeriodRepo apRepo.ActivityPeriodRepository,
	activityRepo actRepo.ActivityRepository,
	periodRepo periodRepo.AcademicPeriodRepository,
) *ListActivityPeriodsUseCase {
	return &ListActivityPeriodsUseCase{activityPeriodRepo: activityPeriodRepo, activityRepo: activityRepo, periodRepo: periodRepo}
}

func (uc *ListActivityPeriodsUseCase) Execute(ctx context.Context, q dto.ActivityPeriodListQuery) ([]dto.ActivityPeriodResponse, *dto.Meta, error) {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.Limit == 0 {
		q.Limit = 20
	}

	result, err := uc.activityPeriodRepo.List(ctx, apRepo.ActivityPeriodListQuery{
		ActivityID:       q.ActivityID,
		AcademicPeriodID: q.AcademicPeriodID,
		Status:           q.Status,
		Page:             q.Page,
		Limit:            q.Limit,
	})
	if err != nil {
		return nil, nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	activityMap, periodMap := uc.enrich(ctx, result.Items)

	items := make([]dto.ActivityPeriodResponse, len(result.Items))
	for i, p := range result.Items {
		resp := command.MapActivityPeriodToResponse(p)
		if a, ok := activityMap[p.ActivityID]; ok {
			resp.ActivityCode = a.Code
			resp.ActivityName = a.Name
		}
		if name, ok := periodMap[p.AcademicPeriodID]; ok {
			resp.PeriodName = name
		}
		items[i] = *resp
	}
	return items, dto.NewMeta(q.Page, q.Limit, result.Total), nil
}

func (uc *ListActivityPeriodsUseCase) enrich(ctx context.Context, items []*apEntity.ActivityPeriod) (map[string]*actEntity.Activity, map[string]string) {
	activityIDs := map[string]struct{}{}
	periodIDs := map[string]struct{}{}
	for _, p := range items {
		activityIDs[p.ActivityID] = struct{}{}
		periodIDs[p.AcademicPeriodID] = struct{}{}
	}

	activityMap := map[string]*actEntity.Activity{}
	if acts, err := uc.activityRepo.FindByIDs(ctx, idsFromSet(activityIDs)); err == nil {
		for _, a := range acts {
			activityMap[a.ID] = a
		}
	}

	periodMap := map[string]string{}
	if periods, err := uc.periodRepo.FindByIDs(ctx, idsFromSet(periodIDs)); err == nil {
		for _, p := range periods {
			periodMap[p.ID] = p.Name
		}
	}
	return activityMap, periodMap
}

func idsFromSet(ids map[string]struct{}) []string {
	out := make([]string, 0, len(ids))
	for id := range ids {
		out = append(out, id)
	}
	return out
}
