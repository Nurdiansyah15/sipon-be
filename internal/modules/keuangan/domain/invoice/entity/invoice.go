package entity

import (
	"time"

	"sipon-be/internal/modules/keuangan/domain/invoice/constant"
	"sipon-be/internal/shared/kernel"
)

type Invoice struct {
	ID              string
	InvoiceNumber   string
	SantriID        string
	UserID          string
	BillingSchemeID *string
	FeeComponentID  string
	BillingPeriodID *string
	Amount          float64
	DiscountAmount  float64
	PaidAmount      float64
	Status          constant.InvoiceStatus
	DueDate         time.Time
	IssuedAt        *time.Time
	Notes           *string
	CreatedBy       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

func NewInvoice(id, invoiceNumber, santriID, userID, feeComponentID string, billingPeriodID *string, amount float64, dueDate time.Time, createdBy string) (*Invoice, error) {
	if id == "" || invoiceNumber == "" || santriID == "" || userID == "" || feeComponentID == "" || createdBy == "" {
		return nil, kernel.WrapMsg(constant.CodeInvoiceNotFound, "Data invoice tidak lengkap", nil)
	}
	now := time.Now()
	return &Invoice{
		ID:              id,
		InvoiceNumber:   invoiceNumber,
		SantriID:        santriID,
		UserID:          userID,
		FeeComponentID:  feeComponentID,
		BillingPeriodID: billingPeriodID,
		Amount:          amount,
		Status:          constant.StatusDraft,
		DueDate:         dueDate,
		CreatedBy:       createdBy,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (i *Invoice) Issue(issuedDate time.Time) error {
	if i.Status != constant.StatusDraft {
		return kernel.WrapMsg(constant.CodeInvoiceInvalidStatus, "Hanya invoice berstatus draft yang dapat diterbitkan", nil)
	}
	i.Status = constant.StatusIssued
	i.IssuedAt = &issuedDate
	i.UpdatedAt = time.Now()
	return nil
}

func (i *Invoice) ApplyDiscount(amount float64) {
	i.DiscountAmount += amount
	i.UpdatedAt = time.Now()
}

func (i *Invoice) AddPayment(amount float64) error {
	if i.Status != constant.StatusIssued && i.Status != constant.StatusPartial {
		return kernel.WrapMsg(constant.CodeInvoiceInvalidStatus, "Invoice tidak dalam status yang dapat menerima pembayaran", nil)
	}
	i.PaidAmount += amount
	i.UpdatedAt = time.Now()
	if i.PaidAmount >= (i.Amount - i.DiscountAmount) {
		i.Status = constant.StatusPaid
	} else {
		i.Status = constant.StatusPartial
	}
	return nil
}

func (i *Invoice) Expire() error {
	if i.Status != constant.StatusIssued && i.Status != constant.StatusPartial {
		return kernel.WrapMsg(constant.CodeInvoiceInvalidStatus, "Hanya invoice yang diterbitkan yang dapat kedaluwarsa", nil)
	}
	i.Status = constant.StatusExpired
	i.UpdatedAt = time.Now()
	return nil
}

func (i *Invoice) Cancel() error {
	if i.PaidAmount > 0 || i.Status == constant.StatusCancelled {
		return kernel.WrapMsg(constant.CodeInvoiceInvalidStatus, "Invoice dengan pembayaran atau yang sudah dibatalkan tidak dapat dibatalkan", nil)
	}
	i.Status = constant.StatusCancelled
	i.UpdatedAt = time.Now()
	return nil
}

func (i *Invoice) HasOutstanding() bool {
	return i.Status == constant.StatusIssued || i.Status == constant.StatusPartial || i.Status == constant.StatusExpired
}

func (i *Invoice) RemainingAmount() float64 {
	remaining := i.Amount - i.DiscountAmount - i.PaidAmount
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (i *Invoice) SoftDelete() {
	now := time.Now()
	i.DeletedAt = &now
	i.UpdatedAt = now
}
