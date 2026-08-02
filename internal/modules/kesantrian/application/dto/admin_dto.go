package dto

import "time"

type CreateSantriRequest struct {
	NIS string `json:"nis" binding:"required"`
}

type CreateSantriResponse struct {
	UserID            string `json:"user_id"`
	SantriID          string `json:"santri_id"`
	NIS               string `json:"nis"`
	GeneratedPassword string `json:"generated_password"`
}

type ListSantriQuery struct {
	NIS      string `form:"nis"`
	SortBy   string `form:"sort_by"`
	SortType string `form:"sort_type"`
	PaginationParams
}

type ListSantriItem struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	NIS       *string   `json:"nis,omitempty"`
	Fullname  *string   `json:"fullname,omitempty"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
