package repository

import (
	"context"

	"sipon-be/internal/modules/akademik/domain/activity_period/entity"
)

type ActivityPeriodListQuery struct {
	ActivityID       *string
	AcademicPeriodID *string
	Status           *string
	Page             int
	Limit            int
}

type ActivityPeriodListResult struct {
	Items []*entity.ActivityPeriod
	Total int64
}

type ActivityPeriodRepository interface {
	Save(ctx context.Context, period *entity.ActivityPeriod) error
	Update(ctx context.Context, period *entity.ActivityPeriod) error
	FindByID(ctx context.Context, id string) (*entity.ActivityPeriod, error)
	FindByActivityAndPeriod(ctx context.Context, activityID, academicPeriodID string) (*entity.ActivityPeriod, error)
	FindByIDs(ctx context.Context, ids []string) ([]*entity.ActivityPeriod, error)
	// ListByPeriodAndProgram returns active activity periods for the given
	// academic period that apply to programID (no program scope at all, or
	// scope explicitly includes programID).
	ListByPeriodAndProgram(ctx context.Context, periodID, programID string) ([]*entity.ActivityPeriod, error)
	List(ctx context.Context, query ActivityPeriodListQuery) (*ActivityPeriodListResult, error)
}
