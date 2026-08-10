package entity

import (
	"time"

	"sipon-be/internal/modules/feedback/domain/like/constant"
)

type Like struct {
	ID         string
	UserID     string
	TargetType constant.LikeTargetType
	TargetID   string
	CreatedAt  time.Time
}

func NewLike(id, userID string, targetType constant.LikeTargetType, targetID string) *Like {
	return &Like{
		ID:         id,
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
		CreatedAt:  time.Now(),
	}
}
