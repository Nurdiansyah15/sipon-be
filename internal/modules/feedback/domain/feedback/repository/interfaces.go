package repository

import (
	"context"

	"sipon-be/internal/modules/feedback/domain/feedback/entity"
)

type FeedbackListQuery struct {
	Category        *string
	Search          string
	UserID          string
	IncludeTakedown bool
	Page            int
	Limit           int
}

type FeedbackListResult struct {
	Items []*entity.Feedback
	Total int64
}

type FeedbackRepository interface {
	Save(ctx context.Context, f *entity.Feedback) error
	Update(ctx context.Context, f *entity.Feedback) error
	FindByID(ctx context.Context, id string) (*entity.Feedback, error)
	List(ctx context.Context, q FeedbackListQuery) (*FeedbackListResult, error)
	IncrementLikeCount(ctx context.Context, id string) error
	DecrementLikeCount(ctx context.Context, id string) error
	IncrementCommentCount(ctx context.Context, id string) error
	DecrementCommentCount(ctx context.Context, id string) error
}
