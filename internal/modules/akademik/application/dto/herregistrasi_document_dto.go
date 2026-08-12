package dto

import "time"

// --- Blueprint dokumen herregistrasi ---

type CreateHerregistrasiDocumentRequirementRequest struct {
	Kind        string  `json:"kind" binding:"required"`
	Label       string  `json:"label" binding:"required"`
	IsRequired  *bool   `json:"is_required"`
	Description *string `json:"description,omitempty"`
}

type UpdateHerregistrasiDocumentRequirementRequest struct {
	Label       *string `json:"label,omitempty"`
	IsRequired  *bool   `json:"is_required,omitempty"`
	Description *string `json:"description,omitempty"`
}

type HerregistrasiDocumentRequirementResponse struct {
	ID               string     `json:"id"`
	AcademicPeriodID string     `json:"academic_period_id"`
	Kind             string     `json:"kind"`
	Label            string     `json:"label"`
	IsRequired       bool       `json:"is_required"`
	Description      *string    `json:"description,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// --- Dokumen herregistrasi ---

type HerregistrasiDocumentPresignRequest struct {
	Kind        string `json:"kind" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
	Filename    string `json:"filename" binding:"required"`
}

type HerregistrasiDocumentPresignResponse struct {
	PresignURL string `json:"presign_url"`
	Key        string `json:"key"`
	PublicURL  string `json:"public_url,omitempty"`
	ExpiresIn  int    `json:"expires_in"`
}

type HerregistrasiDocumentConfirmRequest struct {
	Key              string `json:"key" binding:"required"`
	Kind             string `json:"kind" binding:"required"`
	OriginalFilename string `json:"original_filename"`
	MimeType         string `json:"mime_type"`
	Size             int64  `json:"size"`
}

type HerregistrasiDocumentResponse struct {
	ID                   string     `json:"id"`
	SantriRegistrationID string     `json:"santri_registration_id"`
	Kind                 string     `json:"kind"`
	KindLabel            string     `json:"kind_label,omitempty"`
	Key                  string     `json:"key"`
	OriginalFilename     *string    `json:"original_filename,omitempty"`
	MimeType             *string    `json:"mime_type,omitempty"`
	Size                 *int64     `json:"size,omitempty"`
	Status               string     `json:"status"`
	Notes                *string    `json:"notes,omitempty"`
	VerifiedBy           *string    `json:"verified_by,omitempty"`
	VerifiedAt           *time.Time `json:"verified_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// --- Review & revisi ---

type DokumenVerifyRequest struct{}

type DokumenRejectRequest struct {
	Notes string `json:"notes" binding:"required"`
}

type RevisionRequest struct {
	Notes string `json:"notes" binding:"required"`
}

type HerregistrasiDocumentDownloadResponse struct {
	DownloadURL string `json:"download_url"`
	ExpiresIn   int    `json:"expires_in"`
}

// MyHerregistrasiDetailResponse adalah detail herreg santri pada periode
// aktif: status herreg, blueprint dokumen, dan dokumen yang sudah di-upload.
type MyHerregistrasiDetailResponse struct {
	AcademicPeriod *AcademicPeriodResponse                     `json:"academic_period"`
	Registration   *SantriRegistrationResponse                 `json:"registration,omitempty"`
	Requirements   []HerregistrasiDocumentRequirementResponse  `json:"requirements"`
	Documents      []HerregistrasiDocumentResponse             `json:"documents"`
}
