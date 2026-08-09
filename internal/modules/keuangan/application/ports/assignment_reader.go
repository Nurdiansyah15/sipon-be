package ports

import (
	"context"
	"time"
)

type AssignmentReadModel struct {
	ID              string
	SantriID        string
	BillingSchemeID string
	EffectiveFrom   time.Time
	EffectiveUntil  *time.Time
	AssignedBy      string
	CreatedAt       time.Time
}

type AssignmentReader interface {
	ListActive(ctx context.Context) ([]AssignmentReadModel, error)
}
