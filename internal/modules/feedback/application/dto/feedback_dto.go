package dto

import "time"

type UserSummaryDTO struct {
	UserID   string  `json:"user_id"`
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Fullname *string `json:"fullname,omitempty"`
}

type CreateFeedbackRequest struct {
	Title    string `json:"title" binding:"required"`
	Body     string `json:"body" binding:"required"`
	Category string `json:"category"`
}

type UpdateFeedbackRequest struct {
	Title    string `json:"title" binding:"required"`
	Body     string `json:"body" binding:"required"`
	Category string `json:"category"`
}

type ListFeedbackQuery struct {
	Category string `form:"category"`
	Search   string `form:"search"`
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
}

type ListFeedbackItem struct {
	ID              string          `json:"id"`
	User            *UserSummaryDTO `json:"user,omitempty"`
	Title           string          `json:"title"`
	Body            string          `json:"body"`
	Category        string          `json:"category"`
	IsTakedown      bool            `json:"is_takedown"`
	TakedownReason  *string         `json:"takedown_reason,omitempty"`
	LikeCount       int             `json:"like_count"`
	CommentCount    int             `json:"comment_count"`
	IsLiked         bool            `json:"is_liked"`
	AttachmentCount int             `json:"attachment_count"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type FeedbackDetailResponse struct {
	ListFeedbackItem
	Attachments []AttachmentResponse `json:"attachments"`
	IsOwner     bool                 `json:"is_owner"`
}

type PaginationMeta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	TotalPages  int `json:"total_pages"`
}
