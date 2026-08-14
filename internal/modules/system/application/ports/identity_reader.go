package ports

import (
	"context"

	"sipon-be/internal/modules/identity"
)

// IdentityReader adalah port milik module system untuk membaca scope yang
// dibawa user lewat role-nya. Implementasi (identitygateway) memanggil
// identity.Contract — module system tidak pernah menyentuh domain identity.
// Lihat docs/architecture/module-boundaries.md.
type IdentityReader interface {
	// GetUserScopeSet mengembalikan scope efektif user + penanda apakah user
	// memegang role superuser (system).
	GetUserScopeSet(ctx context.Context, userID string) (*identity.UserScopeSet, error)
}
