package ports

import (
	"context"
	"time"
)

type InvoiceSummaryReadModel struct {
	BillingPeriodID   string
	BillingPeriodName string
	TotalTagihan      float64
	TotalTerbayar     float64
	TotalTunggakan    float64
	JumlahInvoice     int64
	JumlahLunas       int64
	JumlahBelum       int64
}

type OutstandingReadModel struct {
	SantriID         string
	TotalOutstanding float64
	JumlahInvoice    int
}

type LedgerLineReadModel struct {
	Date          time.Time
	JournalNumber string
	Description   string
	Debit         float64
	Credit        float64
}

type AccountBalanceReadModel struct {
	AccountID string
	Debit     float64
	Credit    float64
}

type ReportReader interface {
	InvoiceSummary(ctx context.Context, billingPeriodID, periodID *string) ([]InvoiceSummaryReadModel, error)
	OutstandingBySantri(ctx context.Context, billingPeriodID, periodID *string, page, limit int) ([]OutstandingReadModel, int64, error)
	LedgerLines(ctx context.Context, accountID string, from, to time.Time) ([]LedgerLineReadModel, error)
	BalanceBefore(ctx context.Context, accountID string, before time.Time) (debit, credit float64, err error)
	AccountBalancesToDate(ctx context.Context, asOfDate *string) ([]AccountBalanceReadModel, error)
	AccountBalancesByPeriod(ctx context.Context, periodID string) ([]AccountBalanceReadModel, error)
}
