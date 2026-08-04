package dto

import "time"

type DokumenAsetPresignRequest struct {
	ContentType string `json:"content_type" binding:"required"`
	Filename    string `json:"filename" binding:"required"`
	Kategori    string `json:"kategori" binding:"required"`
	Deskripsi   string `json:"deskripsi,omitempty"`
	IsPublic    bool   `json:"is_public"`
}

type DokumenAsetPresignResponse struct {
	PresignURL string `json:"presign_url"`
	Key        string `json:"key"`
	ExpiresIn  int    `json:"expires_in"`
}

type DokumenAsetConfirmRequest struct {
	Key              string `json:"key" binding:"required"`
	Judul            string `json:"judul" binding:"required"`
	Kategori         string `json:"kategori" binding:"required"`
	Deskripsi        string `json:"deskripsi,omitempty"`
	OriginalFilename string `json:"original_filename" binding:"required"`
	MimeType         string `json:"mime_type" binding:"required"`
	Size             int64  `json:"size" binding:"required"`
	IsPublic         bool   `json:"is_public"`
}

type DokumenAsetConfirmResponse struct {
	ID        string    `json:"id"`
	Judul     string    `json:"judul"`
	Kategori  string    `json:"kategori"`
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"created_at"`
}

type DokumenAsetUpdateRequest struct {
	Judul     *string `json:"judul,omitempty"`
	Deskripsi *string `json:"deskripsi,omitempty"`
	Kategori  *string `json:"kategori,omitempty"`
	IsPublic  *bool   `json:"is_public,omitempty"`
}

type DokumenAsetItem struct {
	ID        string     `json:"id"`
	Judul     string     `json:"judul"`
	Deskripsi *string    `json:"deskripsi,omitempty"`
	Kategori  string     `json:"kategori"`
	Filename  string     `json:"filename"`
	MimeType  string     `json:"mime_type"`
	Size      int64      `json:"size"`
	IsPublic  bool       `json:"is_public"`
	CreatedBy string     `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type DokumenAsetDetail struct {
	DokumenAsetItem
	DownloadURL *string `json:"download_url,omitempty"`
}

type DokumenAsetDownloadResponse struct {
	AccessURL string `json:"access_url"`
	ExpiresIn int    `json:"expires_in"`
}

type DokumenAsetListQuery struct {
	Kategori  string `form:"kategori"`
	Search    string `form:"search"`
	Page      int    `form:"page"`
	Limit     int    `form:"limit"`
}

type DokumenAsetMeta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	TotalPages  int `json:"total_pages"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
