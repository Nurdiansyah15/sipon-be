package repository

import (
	"context"

	"sipon-be/internal/modules/feedback/domain/like/constant"
	"sipon-be/internal/modules/feedback/domain/like/entity"
)

type LikeRepository interface {
	Save(ctx context.Context, l *entity.Like) error
	Delete(ctx context.Context, userID string, targetType constant.LikeTargetType, targetID string) error
	Exists(ctx context.Context, userID string, targetType constant.LikeTargetType, targetID string) (bool, error)
	ListLikedTargetIDs(ctx context.Context, userID string, targetType constant.LikeTargetType, targetIDs []string) (map[string]bool, error)
}
