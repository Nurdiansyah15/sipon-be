package entity

import (
	"time"

	"sipon-be/internal/modules/keuangan/domain/payment/constant"
	"sipon-be/internal/shared/kernel"
)

type Payment struct {
	ID              string
	PaymentNumber   string
	InvoiceID       string
	DebitAccountID  *string
	Amount          float64
	Method          constant.PaymentMethod
	ReferenceNumber *string
	PaymentDate     time.Time
	Status          constant.PaymentStatus
	VerifiedBy      *string
	VerifiedAt      *time.Time
	Notes           *string
	ProofKey        *string
	CreatedBy       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewPayment(id, paymentNumber, invoiceID string, amount float64, method constant.PaymentMethod, paymentDate time.Time, debitAccountID *string, referenceNumber *string, notes *string, proofKey *string, createdBy string) (*Payment, error) {
	if id == "" || paymentNumber == "" || invoiceID == "" || createdBy == "" {
		return nil, kernel.New(constant.CodePaymentNotFound)
	}
	if amount <= 0 {
		return nil, kernel.New(constant.CodePaymentNotFound)
	}
	now := time.Now()
	return &Payment{
		ID:              id,
		PaymentNumber:   paymentNumber,
		InvoiceID:       invoiceID,
		DebitAccountID:  debitAccountID,
		Amount:          amount,
		Method:          method,
		ReferenceNumber: referenceNumber,
		PaymentDate:     paymentDate,
		Status:          constant.PaymentPending,
		Notes:           notes,
		ProofKey:        proofKey,
		CreatedBy:       createdBy,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (p *Payment) Verify(verifierID string) error {
	if p.Status != constant.PaymentPending {
		return kernel.New(constant.CodePaymentInvalidStatus)
	}
	now := time.Now()
	p.Status = constant.PaymentVerified
	p.VerifiedBy = &verifierID
	p.VerifiedAt = &now
	p.UpdatedAt = now
	return nil
}

func (p *Payment) Reject() error {
	if p.Status != constant.PaymentPending {
		return kernel.New(constant.CodePaymentInvalidStatus)
	}
	p.Status = constant.PaymentRejected
	p.UpdatedAt = time.Now()
	return nil
}
