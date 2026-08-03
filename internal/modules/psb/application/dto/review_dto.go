package dto

import "time"

type ReviewResponse struct {
	ID          string    `json:"id"`
	PendaftarID string    `json:"pendaftar_id"`
	Stage       string    `json:"stage"`
	Action      string    `json:"action"`
	Notes       *string   `json:"notes,omitempty"`
	ReviewedBy  string    `json:"reviewed_by"`
	CreatedAt   time.Time `json:"created_at"`
}
