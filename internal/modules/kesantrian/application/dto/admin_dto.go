package dto

import "time"

type CreateSantriRequest struct {
	NIS string `json:"nis" binding:"required"`
	// Gender ('1' laki-laki / '2' perempuan) opsional; bila diberikan harus
	// cocok dengan digit gender pada NIS (karakter ke-5).
	Gender *string `json:"gender,omitempty"`
	// ProgramID optional; bila kosong akan memakai default program dari
	// pengaturan akademik.
	ProgramID *string `json:"program_id,omitempty"`
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
