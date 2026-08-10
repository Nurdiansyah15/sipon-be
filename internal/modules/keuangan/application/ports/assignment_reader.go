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
	// ListAll mengembalikan seluruh penugasan (aktif maupun riwayat). Bila
	// santriID tidak nil, hasil difilter ke satu santri.
	ListAll(ctx context.Context, santriID *string) ([]AssignmentReadModel, error)
}
