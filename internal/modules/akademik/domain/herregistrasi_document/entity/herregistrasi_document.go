package entity

import (
	"time"

	"sipon-be/internal/modules/akademik/domain/herregistrasi_document/constant"
	"sipon-be/internal/shared/kernel"
)

// HerregistrasiDocument adalah dokumen yang di-upload santri untuk suatu
// herregistrasi. Satu dokumen per (registration_id, kind).
type HerregistrasiDocument struct {
	ID                   string
	SantriRegistrationID string
	Kind                 string
	Key                  string
	Status               constant.HerregistrasiDocumentStatus

	OriginalFilename *string
	MimeType         *string
	Size             *int64
	Notes            *string

	VerifiedBy *string
	VerifiedAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func NewHerregistrasiDocument(id, registrationID, kind, key string) (*HerregistrasiDocument, error) {
	if id == "" || registrationID == "" || kind == "" || key == "" {
		return nil, kernel.New(constant.CodeHerregistrasiDocumentInvalidRegistration)
	}
	now := time.Now()
	return &HerregistrasiDocument{
		ID:                   id,
		SantriRegistrationID: registrationID,
		Kind:                 kind,
		Key:                  key,
		Status:               constant.HerregistrasiDocumentStatusPending,
		CreatedAt:            now,
		UpdatedAt:            now,
	}, nil
}

func (d *HerregistrasiDocument) SetMetadata(originalFilename, mimeType string, size int64) {
	d.OriginalFilename = &originalFilename
	d.MimeType = &mimeType
	d.Size = &size
	d.UpdatedAt = time.Now()
}

func (d *HerregistrasiDocument) Verify(verifierID string) error {
	if d.Status == constant.HerregistrasiDocumentStatusVerified {
		return nil
	}
	if d.Status == constant.HerregistrasiDocumentStatusRejected {
		return kernel.New(constant.CodeHerregistrasiDocumentInvalidStatus)
	}
	now := time.Now()
	d.Status = constant.HerregistrasiDocumentStatusVerified
	d.VerifiedBy = &verifierID
	d.VerifiedAt = &now
	d.Notes = nil
	d.UpdatedAt = now
	return nil
}

func (d *HerregistrasiDocument) Reject(verifierID string, notes *string) error {
	now := time.Now()
	d.Status = constant.HerregistrasiDocumentStatusRejected
	d.VerifiedBy = &verifierID
	d.VerifiedAt = &now
	d.Notes = notes
	d.UpdatedAt = now
	return nil
}

func (d *HerregistrasiDocument) SoftDelete() {
	now := time.Now()
	d.DeletedAt = &now
	d.UpdatedAt = now
}
