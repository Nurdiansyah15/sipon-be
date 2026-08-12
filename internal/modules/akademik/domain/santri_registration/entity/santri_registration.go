package entity

import (
	"time"

	"sipon-be/internal/modules/akademik/domain/santri_registration/constant"
	"sipon-be/internal/shared/kernel"
)

type SantriRegistration struct {
	ID               string
	SantriID         string
	AcademicPeriodID string
	Status           constant.SantriRegistrationStatus
	RegisteredAt     *time.Time
	RevisionNotes    *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

func NewSantriRegistration(id, santriID, academicPeriodID string) (*SantriRegistration, error) {
	if id == "" || santriID == "" || academicPeriodID == "" {
		return nil, kernel.New(constant.CodeSantriRegistrationNotFound)
	}
	now := time.Now()
	return &SantriRegistration{
		ID:               id,
		SantriID:         santriID,
		AcademicPeriodID: academicPeriodID,
		Status:           constant.SantriRegistrationStatusPending,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

// NewDraftSantriRegistration membuat herreg berstatus draft — record awal
// agar santri bisa mengisi dokumen sebelum mengajukan.
func NewDraftSantriRegistration(id, santriID, academicPeriodID string) (*SantriRegistration, error) {
	if id == "" || santriID == "" || academicPeriodID == "" {
		return nil, kernel.New(constant.CodeSantriRegistrationNotFound)
	}
	now := time.Now()
	return &SantriRegistration{
		ID:               id,
		SantriID:         santriID,
		AcademicPeriodID: academicPeriodID,
		Status:           constant.SantriRegistrationStatusDraft,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

// Submit mengajukan herreg draft sehingga masuk antrian review admin.
func (r *SantriRegistration) Submit() error {
	if r.Status != constant.SantriRegistrationStatusDraft {
		return kernel.New(constant.CodeSantriRegistrationInvalidStatus)
	}
	r.Status = constant.SantriRegistrationStatusPending
	r.UpdatedAt = time.Now()
	return nil
}

func (r *SantriRegistration) Complete() error {
	if r.Status != constant.SantriRegistrationStatusPending &&
		r.Status != constant.SantriRegistrationStatusRevision {
		return kernel.New(constant.CodeSantriRegistrationInvalidStatus)
	}
	now := time.Now()
	r.Status = constant.SantriRegistrationStatusCompleted
	r.RegisteredAt = &now
	r.RevisionNotes = nil
	r.UpdatedAt = now
	return nil
}

func (r *SantriRegistration) Cancel() error {
	if r.Status != constant.SantriRegistrationStatusPending &&
		r.Status != constant.SantriRegistrationStatusRevision {
		return kernel.New(constant.CodeSantriRegistrationInvalidStatus)
	}
	r.Status = constant.SantriRegistrationStatusCancelled
	r.UpdatedAt = time.Now()
	return nil
}

// RequestRevision menandai herregistrasi perlu revisi dan mencatat catatan
// dari admin. Hanya berlaku dari status pending atau revision.
func (r *SantriRegistration) RequestRevision(notes string) error {
	if r.Status != constant.SantriRegistrationStatusPending &&
		r.Status != constant.SantriRegistrationStatusRevision {
		return kernel.New(constant.CodeSantriRegistrationInvalidStatus)
	}
	now := time.Now()
	r.Status = constant.SantriRegistrationStatusRevision
	r.RevisionNotes = &notes
	r.UpdatedAt = now
	return nil
}

func (r *SantriRegistration) SoftDelete() {
	now := time.Now()
	r.DeletedAt = &now
	r.UpdatedAt = now
}
