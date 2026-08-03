package entity

import (
	"time"

	"sipon-be/internal/modules/psb/domain/dokumen/constant"
	"sipon-be/internal/shared/kernel"
)

type PendaftarDokumen struct {
	ID           string
	PendaftarID  string
	Stage        constant.DokumenStage
	Kind         constant.DokumenKind
	Key          string
	Status       constant.DokumenStatus

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

func NewPendaftarDokumen(id, pendaftarID string, stage constant.DokumenStage, kind constant.DokumenKind, key string) (*PendaftarDokumen, error) {
	if !constant.ValidDokumenKinds[kind] {
		return nil, kernel.New(constant.CodeDokumenInvalidKind)
	}
	now := time.Now()
	return &PendaftarDokumen{
		ID:          id,
		PendaftarID: pendaftarID,
		Stage:       stage,
		Kind:        kind,
		Key:         key,
		Status:      constant.DokumenStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (d *PendaftarDokumen) Verify(verifierID string) error {
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

func (d *PendaftarDokumen) Reject(verifierID string, notes *string) error {
	now := time.Now()
	d.Status = constant.DokumenStatusRejected
	d.VerifiedBy = &verifierID
	d.VerifiedAt = &now
	d.Notes = notes
	d.UpdatedAt = now
	return nil
}

func (d *PendaftarDokumen) SoftDelete() {
	now := time.Now()
	d.DeletedAt = &now
	d.UpdatedAt = now
}
