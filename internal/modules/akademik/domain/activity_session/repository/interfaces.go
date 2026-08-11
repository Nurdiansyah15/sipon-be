package repository

import (
	"context"

	"sipon-be/internal/modules/akademik/domain/activity_session/entity"
)

type ActivitySessionListQuery struct {
	ActivityScheduleID *string
	AcademicPeriodID   *string
	Status             *string
	StartDate          *string
	EndDate            *string
	Page               int
	Limit              int
}

type ActivitySessionListResult struct {
	Items []*entity.ActivitySession
	Total int64
}

type ActivitySessionRepository interface {
	Save(ctx context.Context, session *entity.ActivitySession) error
	Update(ctx context.Context, session *entity.ActivitySession) error
	FindByID(ctx context.Context, id string) (*entity.ActivitySession, error)
	List(ctx context.Context, query ActivitySessionListQuery) (*ActivitySessionListResult, error)
}
