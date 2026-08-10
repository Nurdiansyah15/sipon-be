package repository

import (
	"context"

	"sipon-be/internal/modules/feedback/domain/comment/entity"
)

type CommentListQuery struct {
	FeedbackID      string
	IncludeTakedown bool
	Page            int
	Limit           int
}

type CommentListResult struct {
	Items []*entity.Comment
	Total int64
}

type CommentRepository interface {
	Save(ctx context.Context, c *entity.Comment) error
	Update(ctx context.Context, c *entity.Comment) error
	FindByID(ctx context.Context, id string) (*entity.Comment, error)
	List(ctx context.Context, q CommentListQuery) (*CommentListResult, error)
	IncrementLikeCount(ctx context.Context, id string) error
	DecrementLikeCount(ctx context.Context, id string) error
}
