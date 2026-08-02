package dto

import "time"

type DokumenPresignRequest struct {
	ContentType string `json:"content_type" binding:"required"`
	Kind        string `json:"kind" binding:"required"`
}

type DokumenPresignResponse struct {
	PresignURL string `json:"presign_url"`
	Key        string `json:"key"`
	ExpiresIn  int    `json:"expires_in"`
}

type DokumenConfirmRequest struct {
	Kind             string  `json:"kind" binding:"required"`
	Key              string  `json:"key" binding:"required"`
	OriginalFilename *string `json:"original_filename,omitempty"`
	MimeType         *string `json:"mime_type,omitempty"`
	Size             *int64  `json:"size,omitempty"`
}

type DokumenConfirmResponse struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	Key       string    `json:"key"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type DokumenItem struct {
	ID               string     `json:"id"`
	Kind             string     `json:"kind"`
	Key              string     `json:"key"`
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
	AccessURL string `json:"access_url"`
	ExpiresIn int    `json:"expires_in"`
}

type RejectDokumenRequest struct {
	Notes *string `json:"notes,omitempty"`
}
