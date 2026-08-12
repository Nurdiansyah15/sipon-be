package entity

import (
	"time"

	"sipon-be/internal/modules/akademik/domain/herregistrasi_document_requirement/constant"
	"sipon-be/internal/shared/kernel"
)

// HerregistrasiDocumentRequirement adalah blueprint dokumen yang boleh
// di-upload santri pada suatu periode akademik (herregistrasi).
type HerregistrasiDocumentRequirement struct {
	ID               string
	AcademicPeriodID string
	Kind             string
	Label            string
	IsRequired       bool
	Description      *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

func NewHerregistrasiDocumentRequirement(id, academicPeriodID, kind, label string, isRequired bool, description *string) (*HerregistrasiDocumentRequirement, error) {
	if id == "" || academicPeriodID == "" || kind == "" || label == "" {
		return nil, kernel.New(constant.CodeHerregistrasiDocumentRequirementInvalidKind)
	}
	now := time.Now()
	return &HerregistrasiDocumentRequirement{
		ID:               id,
		AcademicPeriodID: academicPeriodID,
		Kind:             kind,
		Label:            label,
		IsRequired:       isRequired,
		Description:      description,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

func (r *HerregistrasiDocumentRequirement) Update(label *string, isRequired *bool, description *string) error {
	if label != nil && *label != "" {
		r.Label = *label
	}
	if isRequired != nil {
		r.IsRequired = *isRequired
	}
	if description != nil {
		r.Description = description
	}
	r.UpdatedAt = time.Now()
	return nil
}

func (r *HerregistrasiDocumentRequirement) SoftDelete() {
	now := time.Now()
	r.DeletedAt = &now
	r.UpdatedAt = now
}
