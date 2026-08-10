package dto

import "time"

type CreateTipeSuratRequest struct {
	Nama string `json:"nama" binding:"required"`
	Kode string `json:"kode" binding:"required"`
}

type UpdateTipeSuratRequest struct {
	Nama string `json:"nama" binding:"required"`
	Kode string `json:"kode" binding:"required"`
}

type UpdateTipeSuratNamaRequest struct {
	Nama string `json:"nama" binding:"required"`
}

type TipeSuratResponse struct {
	ID        string    `json:"id"`
	Nama      string    `json:"nama"`
	Kode      string    `json:"kode"`
	CreatedBy *string   `json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListTipeSuratQuery struct {
	SortBy   string `form:"sort_by"`
	SortType string `form:"sort_type"`
	PaginationParams
}
