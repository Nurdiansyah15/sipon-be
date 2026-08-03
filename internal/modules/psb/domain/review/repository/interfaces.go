package repository

import (
	"context"

	"sipon-be/internal/modules/psb/domain/review/entity"
)

type PendaftarReviewRepository interface {
	Save(ctx context.Context, r *entity.PendaftarReview) error
	FindByPendaftarID(ctx context.Context, pendaftarID string) ([]*entity.PendaftarReview, error)
	HardDeleteByPendaftarID(ctx context.Context, pendaftarID string) (int64, error)
}
