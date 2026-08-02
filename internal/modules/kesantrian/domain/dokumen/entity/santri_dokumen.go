package entity

import (
	"time"

	"sipon-be/internal/modules/kesantrian/domain/dokumen/constant"
	"sipon-be/internal/shared/kernel"
)

type SantriDokumen struct {
	ID       string
	SantriID string
	Kind     constant.DokumenKind
	Key      string
	Status   constant.DokumenStatus

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

func NewSantriDokumen(id, santriID string, kind constant.DokumenKind, key string) (*SantriDokumen, error) {
	if !constant.ValidDokumenKinds[kind] {
		return nil, kernel.New(constant.CodeDokumenInvalidKind)
	}
	now := time.Now()
	return &SantriDokumen{
		ID:        id,
		SantriID:  santriID,
		Kind:      kind,
		Key:       key,
		Status:    constant.DokumenStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Verify is idempotent when already verified, but errors if the document
// was previously rejected — mirrors sipon-api's behavior.
func (d *SantriDokumen) Verify(verifierID string) error {
	if d.Status == constant.DokumenStatusVerified {
		return nil
	}
	if d.Status == constant.DokumenStatusRejected {
		return kernel.New(constant.CodeDokumenInvalidStatus)
	}
	now := time.Now()
	d.Status = constant.DokumenStatusVerified
	d.VerifiedBy = &verifierID
	d.VerifiedAt = &now
	d.Notes = nil
	d.UpdatedAt = now
	return nil
}

// Reject is idempotent when already rejected (notes are refreshed).
func (d *SantriDokumen) Reject(verifierID string, notes *string) error {
	now := time.Now()
	d.Status = constant.DokumenStatusRejected
	d.VerifiedBy = &verifierID
	d.VerifiedAt = &now
	d.Notes = notes
	d.UpdatedAt = now
	return nil
}

func (d *SantriDokumen) SoftDelete() {
	now := time.Now()
	d.DeletedAt = &now
	d.UpdatedAt = now
}
