package repository

import (
	"context"

	"sipon-be/internal/modules/feedback/domain/attachment/entity"
)

type AttachmentRepository interface {
	Save(ctx context.Context, a *entity.Attachment) error
	FindByID(ctx context.Context, id string) (*entity.Attachment, error)
	ListByFeedbackID(ctx context.Context, feedbackID string) ([]*entity.Attachment, error)
	CountByFeedbackID(ctx context.Context, feedbackID string) (int64, error)
	CountByFeedbackIDs(ctx context.Context, feedbackIDs []string) (map[string]int64, error)
	SoftDelete(ctx context.Context, id string) error
	MaxSortOrder(ctx context.Context, feedbackID string) (int, error)
}
