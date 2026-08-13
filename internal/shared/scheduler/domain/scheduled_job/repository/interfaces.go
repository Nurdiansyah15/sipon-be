package repository

import (
	"context"
	"time"

	"sipon-be/internal/shared/scheduler/domain/scheduled_job/entity"
)

type Repository interface {
	Save(ctx context.Context, job *entity.ScheduledJob) error
	FindDueAndClaim(ctx context.Context, now time.Time, limit int) ([]*entity.ScheduledJob, error)
	Update(ctx context.Context, job *entity.ScheduledJob) error
	FindByTypeAndReferenceID(ctx context.Context, jobType string, referenceID string) (*entity.ScheduledJob, error)
}
