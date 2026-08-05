package repository

import (
	"context"

	"sipon-be/internal/modules/keuangan/domain/adjustment/entity"
)

type AdjustmentRepository interface {
	Save(ctx context.Context, adj *entity.InvoiceAdjustment) error
	FindByInvoiceID(ctx context.Context, invoiceID string) ([]*entity.InvoiceAdjustment, error)
}
