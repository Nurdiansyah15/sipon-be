package query

import (
	"context"

	"sipon-be/internal/modules/psb/application"
	"sipon-be/internal/modules/psb/application/dto"
	rrepo "sipon-be/internal/modules/psb/domain/review/repository"
	"sipon-be/internal/shared/kernel"
)

type ListReviewsUseCase struct {
	reviewRepo rrepo.PendaftarReviewRepository
}

func NewListReviewsUseCase(reviewRepo rrepo.PendaftarReviewRepository) *ListReviewsUseCase {
	return &ListReviewsUseCase{reviewRepo: reviewRepo}
}

func (uc *ListReviewsUseCase) Execute(ctx context.Context, pendaftarID string) ([]dto.ReviewResponse, error) {
	reviews, err := uc.reviewRepo.FindByPendaftarID(ctx, pendaftarID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := make([]dto.ReviewResponse, len(reviews))
	for i, r := range reviews {
		items[i] = dto.ReviewResponse{
			ID:          r.ID,
			PendaftarID: r.PendaftarID,
			Stage:       string(r.Stage),
			Action:      string(r.Action),
			Notes:       r.Notes,
			ReviewedBy:  r.ReviewedBy,
			CreatedAt:   r.CreatedAt,
		}
	}

	return items, nil
}
