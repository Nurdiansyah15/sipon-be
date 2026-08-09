package entity

import (
	"time"

	"sipon-be/internal/modules/keuangan/domain/adjustment/constant"
	"sipon-be/internal/shared/kernel"
)

type InvoiceAdjustment struct {
	ID          string
	InvoiceID   string
	Type        constant.AdjustmentType
	Amount      float64
	Percentage  *float64
	Description *string
	AppliedBy   string
	AppliedAt   time.Time
}

func NewInvoiceAdjustment(id, invoiceID string, adjType constant.AdjustmentType, amount float64, percentage *float64, description *string, appliedBy string) (*InvoiceAdjustment, error) {
	if id == "" || invoiceID == "" || appliedBy == "" {
		return nil, kernel.WrapMsg(constant.CodeAdjustmentNotFound, "Data penyesuaian tidak lengkap", nil)
	}
	return &InvoiceAdjustment{
		ID:          id,
		InvoiceID:   invoiceID,
		Type:        adjType,
		Amount:      amount,
		Percentage:  percentage,
		Description: description,
		AppliedBy:   appliedBy,
		AppliedAt:   time.Now(),
	}, nil
}
