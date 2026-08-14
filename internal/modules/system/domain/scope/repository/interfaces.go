package repository

import (
	"context"

	scopesentity "sipon-be/internal/modules/system/domain/scope/entity"
)

type ScopeRepository interface {
	Save(ctx context.Context, scope *scopesentity.Scope) error
	Update(ctx context.Context, scope *scopesentity.Scope) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*scopesentity.Scope, error)
	FindByTypeAndCode(ctx context.Context, scopeType, code string) (*scopesentity.Scope, error)
	ListByType(ctx context.Context, scopeType string, includeInactive bool) ([]*scopesentity.Scope, error)
	ListAll(ctx context.Context, includeInactive bool) ([]*scopesentity.Scope, error)
	// ListActiveCodesByType mengembalikan kode scope aktif untuk scope type
	// tertentu — dipakai resolver akses untuk menetapkan "full access".
	ListActiveCodesByType(ctx context.Context, scopeType string) ([]string, error)
}
