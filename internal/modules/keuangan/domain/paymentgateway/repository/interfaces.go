package repository

import (
	"context"

	"sipon-be/internal/modules/keuangan/domain/paymentgateway/entity"
)

type PaymentGatewayRepository interface {
	Save(ctx context.Context, tx *entity.PaymentGatewayTransaction) error
	Update(ctx context.Context, tx *entity.PaymentGatewayTransaction) error
	FindByTransactionID(ctx context.Context, transactionID string) (*entity.PaymentGatewayTransaction, error)
	FindByInvoiceID(ctx context.Context, invoiceID string) ([]*entity.PaymentGatewayTransaction, error)
	FindActiveByInvoiceID(ctx context.Context, invoiceID string) (*entity.PaymentGatewayTransaction, error)
}
