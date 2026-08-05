package entity

import (
	"time"

	"sipon-be/internal/modules/keuangan/domain/billingscheme/constant"
	"sipon-be/internal/shared/kernel"
)

type BillingSchemeItem struct {
	ID              string
	BillingSchemeID string
	FeeComponentID  string
	AmountOverride  *float64
	IsRequired      bool
	SortOrder       int
	CreatedAt       time.Time
}

func NewBillingSchemeItem(id, billingSchemeID, feeComponentID string, amountOverride *float64, isRequired bool, sortOrder int) (*BillingSchemeItem, error) {
	if id == "" || billingSchemeID == "" || feeComponentID == "" {
		return nil, kernel.New(constant.CodeSchemeItemNotFound)
	}
	return &BillingSchemeItem{
		ID:              id,
		BillingSchemeID: billingSchemeID,
		FeeComponentID:  feeComponentID,
		AmountOverride:  amountOverride,
		IsRequired:      isRequired,
		SortOrder:       sortOrder,
		CreatedAt:       time.Now(),
	}, nil
}

func (i *BillingSchemeItem) GetEffectiveAmount(defaultAmount float64) float64 {
	if i.AmountOverride != nil {
		return *i.AmountOverride
	}
	return defaultAmount
}
