package repository

import (
	"context"

	"sipon-be/internal/modules/keuangan/domain/payment/entity"
	"sipon-be/internal/modules/keuangan/domain/payment/valueobject"
)

type PaymentListQuery struct {
	InvoiceID *string
	Status    *string
	PeriodID  *string
	Page      int
	Limit     int
}

type PaymentListResult struct {
	Items []*entity.Payment
	Total int64
}

type PaymentRepository interface {
	Save(ctx context.Context, p *entity.Payment) error
	Update(ctx context.Context, p *entity.Payment) error
	FindByID(ctx context.Context, id string) (*entity.Payment, error)
	FindByNumber(ctx context.Context, number string) (*entity.Payment, error)
	List(ctx context.Context, query PaymentListQuery) (*PaymentListResult, error)
	FindByInvoiceID(ctx context.Context, invoiceID string) ([]*entity.Payment, error)
	FindVerifiedByInvoiceID(ctx context.Context, invoiceID string) ([]*entity.Payment, error)
	NextPaymentNumber(ctx context.Context) (valueobject.PaymentNumber, error)
}
