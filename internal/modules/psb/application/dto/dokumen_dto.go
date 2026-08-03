package dto

import "time"

type DokumenPresignRequest struct {
	Stage       string `json:"stage" binding:"required"`
	Kind        string `json:"kind" binding:"required"`
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
}

type DokumenPresignResponse struct {
	PresignURL string `json:"presign_url"`
	Key        string `json:"key"`
	PublicURL  string `json:"public_url,omitempty"`
}

type DokumenConfirmRequest struct {
	Stage string `json:"stage" binding:"required"`
	Kind  string `json:"kind" binding:"required"`
	Key   string `json:"key" binding:"required"`
}

type DokumenConfirmResponse struct {
	ID string `json:"id"`
}

type DokumenItemResponse struct {
	ID               string     `json:"id"`
	Stage            string     `json:"stage"`
	Kind             string     `json:"kind"`
	Status           string     `json:"status"`
	OriginalFilename *string    `json:"original_filename,omitempty"`
	MimeType         *string    `json:"mime_type,omitempty"`
	Size             *int64     `json:"size,omitempty"`
	Notes            *string    `json:"notes,omitempty"`
	VerifiedBy       *string    `json:"verified_by,omitempty"`
	VerifiedAt       *time.Time `json:"verified_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

type DokumenAccessResponse struct {
	URL string `json:"url"`
}
