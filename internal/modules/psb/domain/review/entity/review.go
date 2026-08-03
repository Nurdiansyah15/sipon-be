package entity

import (
	"time"

	"sipon-be/internal/modules/psb/domain/review/constant"
	"sipon-be/internal/shared/kernel"
)

type PendaftarReview struct {
	ID           string
	PendaftarID  string
	Stage        constant.ReviewStage
	Action       constant.ReviewAction
	Notes        *string
	ReviewedBy   string
	CreatedAt    time.Time
}

func NewPendaftarReview(id, pendaftarID string, stage constant.ReviewStage, action constant.ReviewAction, reviewedBy string, notes *string) (*PendaftarReview, error) {
	if id == "" || pendaftarID == "" || reviewedBy == "" {
		return nil, kernel.New(constant.ErrCodeInvalidReview)
	}
	return &PendaftarReview{
		ID:          id,
		PendaftarID: pendaftarID,
		Stage:       stage,
		Action:      action,
		Notes:       notes,
		ReviewedBy:  reviewedBy,
		CreatedAt:   time.Now(),
	}, nil
}
