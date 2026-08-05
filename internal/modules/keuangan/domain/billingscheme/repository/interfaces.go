package repository

import (
	"context"
	"time"

	"sipon-be/internal/modules/keuangan/domain/billingscheme/entity"
)

type BillingSchemeListQuery struct {
	Active *bool
	Page   int
	Limit  int
}

type BillingSchemeListResult struct {
	Items []*entity.BillingScheme
	Total int64
}

type BillingSchemeRepository interface {
	Save(ctx context.Context, scheme *entity.BillingScheme) error
	Update(ctx context.Context, scheme *entity.BillingScheme) error
	FindByID(ctx context.Context, id string) (*entity.BillingScheme, error)
	List(ctx context.Context, query BillingSchemeListQuery) (*BillingSchemeListResult, error)
	AddItems(ctx context.Context, schemeID string, items []*entity.BillingSchemeItem) error
	RemoveItem(ctx context.Context, schemeID string, itemID string) error
}

type SantriBillingAssignmentRepository interface {
	Save(ctx context.Context, assignment *entity.SantriBillingAssignment) error
	FindActiveBySantriID(ctx context.Context, santriID string) (*entity.SantriBillingAssignment, error)
	EndAssignment(ctx context.Context, id string, effectiveUntil time.Time) error
}
