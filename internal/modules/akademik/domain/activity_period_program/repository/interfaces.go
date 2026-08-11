package repository

import (
	"context"

	"sipon-be/internal/modules/akademik/domain/activity_period_program/entity"
)

type ActivityPeriodProgramRepository interface {
	Save(ctx context.Context, program *entity.ActivityPeriodProgram) error
	Delete(ctx context.Context, id string) error
	FindByActivityPeriodAndProgram(ctx context.Context, activityPeriodID, programID string) (*entity.ActivityPeriodProgram, error)
	ListByActivityPeriod(ctx context.Context, activityPeriodID string) ([]*entity.ActivityPeriodProgram, error)
}
