package dto

import "time"

type RequestSantriResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type ListSantriRequestsQuery struct {
	Status   string `form:"status"`
	SortBy   string `form:"sort_by"`
	SortType string `form:"sort_type"`
	PaginationParams
}

type SantriRequestItem struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Fullname  *string   `json:"fullname,omitempty"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	Notes     *string   `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ApproveSantriRequestRequest struct {
	NIS string `json:"nis" binding:"required"`
}

type RejectSantriRequestRequest struct {
	Notes *string `json:"notes,omitempty"`
}
