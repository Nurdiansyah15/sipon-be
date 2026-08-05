package keuangan

import (
	"context"
)

type OutstandingSummary struct {
	HasOutstanding   bool
	TotalOutstanding float64
	Count            int
}

type Contract interface {
	GetOutstandingInvoices(ctx context.Context, santriID string) (*OutstandingSummary, error)
	HasPaidComponent(ctx context.Context, santriID, componentCode, periode string) (bool, error)
}

var _ Contract = (*Module)(nil)
