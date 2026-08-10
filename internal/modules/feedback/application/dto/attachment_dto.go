package dto

import "time"

type AttachmentPresignRequest struct {
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
}

type AttachmentPresignResponse struct {
	PresignURL string `json:"presign_url"`
	Key        string `json:"key"`
	PublicURL  string `json:"public_url,omitempty"`
}

type AttachmentConfirmRequest struct {
	Key              string `json:"key" binding:"required"`
	OriginalFilename string `json:"original_filename"`
	MimeType         string `json:"mime_type"`
	Size             int64  `json:"size"`
}

type AttachmentResponse struct {
	ID               string    `json:"id"`
	Key              string    `json:"key"`
	OriginalFilename *string   `json:"original_filename,omitempty"`
	MimeType         *string   `json:"mime_type,omitempty"`
	Size             *int64    `json:"size,omitempty"`
	DownloadURL      string    `json:"download_url,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
