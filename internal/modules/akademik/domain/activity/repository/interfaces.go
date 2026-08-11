package repository

import (
	"context"

	"sipon-be/internal/modules/akademik/domain/activity/entity"
)

type ActivityListQuery struct {
	Status *string
	Search *string
	Page   int
	Limit  int
}

type ActivityListResult struct {
	Items []*entity.Activity
	Total int64
}

type ActivityRepository interface {
	Save(ctx context.Context, activity *entity.Activity) error
	Update(ctx context.Context, activity *entity.Activity) error
	FindByID(ctx context.Context, id string) (*entity.Activity, error)
	FindByCode(ctx context.Context, code string) (*entity.Activity, error)
	FindByIDs(ctx context.Context, ids []string) ([]*entity.Activity, error)
	List(ctx context.Context, query ActivityListQuery) (*ActivityListResult, error)
}
