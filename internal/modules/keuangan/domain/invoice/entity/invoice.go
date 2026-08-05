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
	Periode         string
	TahunAjaran     string
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

func NewInvoice(id, invoiceNumber, santriID, userID, feeComponentID, periode, tahunAjaran string, amount float64, dueDate time.Time, createdBy string) (*Invoice, error) {
	if id == "" || invoiceNumber == "" || santriID == "" || userID == "" || feeComponentID == "" || createdBy == "" {
		return nil, kernel.New(constant.CodeInvoiceNotFound)
	}
	now := time.Now()
	return &Invoice{
		ID:             id,
		InvoiceNumber:  invoiceNumber,
		SantriID:       santriID,
		UserID:         userID,
		FeeComponentID: feeComponentID,
		Periode:        periode,
		TahunAjaran:    tahunAjaran,
		Amount:         amount,
		Status:         constant.StatusDraft,
		DueDate:        dueDate,
		CreatedBy:      createdBy,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func (i *Invoice) Issue() error {
	if i.Status != constant.StatusDraft {
		return kernel.New(constant.CodeInvoiceInvalidStatus)
	}
	now := time.Now()
	i.Status = constant.StatusIssued
	nowDate := now
	i.IssuedAt = &nowDate
	i.UpdatedAt = now
	return nil
}

func (i *Invoice) ApplyDiscount(amount float64) {
	i.DiscountAmount += amount
	i.UpdatedAt = time.Now()
}

func (i *Invoice) AddPayment(amount float64) error {
	if i.Status != constant.StatusIssued && i.Status != constant.StatusPartial {
		return kernel.New(constant.CodeInvoiceInvalidStatus)
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
		return kernel.New(constant.CodeInvoiceInvalidStatus)
	}
	i.Status = constant.StatusExpired
	i.UpdatedAt = time.Now()
	return nil
}

func (i *Invoice) Cancel() error {
	if i.Status == constant.StatusPaid || i.Status == constant.StatusCancelled {
		return kernel.New(constant.CodeInvoiceInvalidStatus)
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
