package entity

import (
	"time"

	"sipon-be/internal/modules/keuangan/domain/billingscheme/constant"
	"sipon-be/internal/shared/kernel"
)

type SantriBillingAssignment struct {
	ID              string
	SantriID        string
	BillingSchemeID string
	EffectiveFrom   time.Time
	EffectiveUntil  *time.Time
	AssignedBy      string
	CreatedAt       time.Time
}

func NewSantriBillingAssignment(id, santriID, billingSchemeID, assignedBy string, effectiveFrom time.Time, effectiveUntil *time.Time) (*SantriBillingAssignment, error) {
	if id == "" || santriID == "" || billingSchemeID == "" || assignedBy == "" {
		return nil, kernel.New(constant.CodeBillingSchemeNotFound)
	}
	return &SantriBillingAssignment{
		ID:              id,
		SantriID:        santriID,
		BillingSchemeID: billingSchemeID,
		EffectiveFrom:   effectiveFrom,
		EffectiveUntil:  effectiveUntil,
		AssignedBy:      assignedBy,
		CreatedAt:       time.Now(),
	}, nil
}
