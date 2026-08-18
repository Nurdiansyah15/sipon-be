package repository

import (
	"context"
	"time"

	"sipon-be/internal/shared/scheduler/domain/scheduled_job/entity"
)

type Repository interface {
	Save(ctx context.Context, job *entity.ScheduledJob) error
	// FindDueAndClaim mengklaim job ACTIVE yang jatuh tempo ATAU job PROCESSING
	// yang lease-nya sudah expired, dengan FOR UPDATE SKIP LOCKED. leaseUntil
	// dipakai untuk mengunci job hasil klaim.
	FindDueAndClaim(ctx context.Context, now time.Time, limit int, leaseUntil time.Time) ([]*entity.ScheduledJob, error)
	Update(ctx context.Context, job *entity.ScheduledJob) error
	FindByTypeAndReferenceID(ctx context.Context, jobType string, referenceID string) (*entity.ScheduledJob, error)
}
