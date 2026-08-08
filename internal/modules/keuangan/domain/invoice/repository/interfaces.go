package repository

import (
	"context"

	"sipon-be/internal/modules/keuangan/domain/invoice/entity"
)

type InvoiceListQuery struct {
	SantriID        *string
	UserID          *string
	Status          *string
	BillingPeriodID *string
	Page            int
	Limit           int
}

type InvoiceListResult struct {
	Items []*entity.Invoice
	Total int64
}

type InvoiceRepository interface {
	Save(ctx context.Context, inv *entity.Invoice) error
	Update(ctx context.Context, inv *entity.Invoice) error
	FindByID(ctx context.Context, id string) (*entity.Invoice, error)
	FindByNumber(ctx context.Context, number string) (*entity.Invoice, error)
	List(ctx context.Context, query InvoiceListQuery) (*InvoiceListResult, error)
	FindBySantriComponentPeriod(ctx context.Context, santriID, feeComponentID, billingPeriodID string) (*entity.Invoice, error)
	FindOutstandingBySantriID(ctx context.Context, santriID string) ([]*entity.Invoice, error)
	HasPaidComponent(ctx context.Context, santriID, componentCode, billingPeriodID string) (bool, error)
	NextInvoiceNumber(ctx context.Context) (string, error)
}
