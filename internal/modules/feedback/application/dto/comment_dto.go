package dto

import "time"

type CreateCommentRequest struct {
	Body      string  `json:"body" binding:"required"`
	ReplyToID *string `json:"reply_to_id,omitempty"`
}

type UpdateCommentRequest struct {
	Body string `json:"body" binding:"required"`
}

type CommentItemResponse struct {
	ID             string          `json:"id"`
	FeedbackID     string          `json:"feedback_id"`
	User           *UserSummaryDTO `json:"user,omitempty"`
	Body           string          `json:"body"`
	ReplyToID      *string         `json:"reply_to_id,omitempty"`
	ReplyToUser    *UserSummaryDTO `json:"reply_to_user,omitempty"`
	IsTakedown     bool            `json:"is_takedown"`
	TakedownReason *string         `json:"takedown_reason,omitempty"`
	LikeCount      int             `json:"like_count"`
	IsLiked        bool            `json:"is_liked"`
	IsOwner        bool            `json:"is_owner"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}
