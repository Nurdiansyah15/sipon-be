package system

import (
	"context"
)

// ScopeInfo adalah DTO kontrak untuk master scope.
type ScopeInfo struct {
	ID        string
	ScopeType string
	Code      string
	Name      string
}

// UserScopeAccess adalah DTO kontrak hasil resolusi akses scope user terhadap
// satu scope type.
type UserScopeAccess struct {
	UserID        string
	ScopeType     string
	HasAccess     bool
	HasFullAccess bool
	// AllowedCodes berisi kode scope yang boleh diakses. Nil/empty saat
	// HasFullAccess true — pemanggil tidak perlu filter tambahan.
	AllowedCodes []string
}

// Contract adalah SATU-SATUNYA surface yang boleh di-import module lain dari
// system. Semua parameter/return value berupa DTO milik system — tidak pernah
// entity domain atau tipe application/ports. Lihat
// docs/architecture/module-boundaries.md.
type Contract interface {
	// GetUserScopeAccess menghitung akses scope user terhadap satu scope type
	// (mis. "gender"). Module pemanggil memakai hasilnya untuk memfilter query
	// resource-nya (IN clause terhadap AllowedCodes, atau tanpa filter saat
	// HasFullAccess).
	GetUserScopeAccess(ctx context.Context, userID, scopeType string) (*UserScopeAccess, error)

	// CanAccessResource mengecek apakah user boleh mengakses sebuah resource
	// yang diklasifikasikan dengan resourceScopeCodes (kode scope master).
	// Resource tanpa scope (list kosong) dianggap publik.
	CanAccessResource(ctx context.Context, userID, scopeType string, resourceScopeCodes []string) (bool, error)
}

var _ Contract = (*Module)(nil)
