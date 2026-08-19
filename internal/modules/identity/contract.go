package identity

import (
	"context"
)

// UserSummary is identity's own contract-boundary DTO — deliberately NOT
// userentity.User and NOT dto.UserManagementResponse (that one is
// admin/query-shaped and will change for reasons unrelated to this contract).
type UserSummary struct {
	UserID    string
	Username  string
	Email     string
	IsActive  bool
	Fullname  *string
	Phone     *string
	AvatarKey *string
}

// CreateAccountInput is the input DTO for CreateAccountWithNIS. NISValue is
// expected to already be validated by the caller's own NIS value object —
// identity does not re-validate NIS format, it only stores the value as a
// login identifier (mirrors sipon-api's identity-side NIS pass-through).
type CreateAccountInput struct {
	Username string
	Email    string
	Fullname *string
	NISValue string
}

// CreateAccountResult carries the one-time plaintext generated password back
// to the caller — it is never persisted or retrievable again after this call.
type CreateAccountResult struct {
	UserID            string
	GeneratedPassword string
}

// Principal mirrors infrastructure/principal.Principal's shape but is its
// own type — the contract must not leak the principal package.
type Principal struct {
	UserID      string
	Roles       []string
	Permissions []string
	Scopes      []ScopeInfo
}

// UserScopeValue adalah satu scope efektif yang dibawa user lewat role-nya
// (mis. scope_type "gender" + scope_value "male").
type UserScopeValue struct {
	ScopeType  string
	ScopeValue string
}

// UserScopeSet adalah himpunan scope efektif user — dipakai module lain
// (mis. system) untuk memfilter data berbasis scope.
type UserScopeSet struct {
	// HasSystemRole true ketika user memegang role superuser (role_type system
	// yang tidak assignable). Module pemanggil boleh memberi akses penuh.
	HasSystemRole bool
	// Values berisi scope unik (deduplicated) lintas seluruh role aktif user.
	Values []UserScopeValue
}

type ScopeInfo struct {
	ScopeType string
	ScopeID   *string
}

// UserScopeAccess adalah DTO kontrak hasil resolusi akses scope user terhadap
// satu scope type (master scope kini dimiliki module identity).
type UserScopeAccess struct {
	UserID        string
	ScopeType     string
	HasAccess     bool
	HasFullAccess bool
	// AllowedCodes berisi kode scope yang boleh diakses. Nil/empty saat
	// HasFullAccess true — pemanggil tidak perlu filter tambahan.
	AllowedCodes []string
}

// UserScopeAccessIDs adalah DTO kontrak hasil resolusi akses scope user yang
// AllowedScopeIDs berisi master scope ID (bukan kode). Dipakai pemanggil yang
// menyimpan scope_id pada resource (mis. kesantrian.surat) untuk tagging &
// filter.
type UserScopeAccessIDs struct {
	UserID        string
	ScopeType     string
	HasAccess     bool
	HasFullAccess bool
	// AllowedScopeIDs berisi master scope ID yang boleh diakses. Nil/empty saat
	// HasFullAccess true.
	AllowedScopeIDs []string
}

// Contract is the ONLY surface another module may import from identity. No
// domain entity, no repository interface, no application/ports type ever
// appears in this signature — only the DTOs declared in this file. See
// docs/architecture/module-boundaries.md for the full cross-module
// isolation convention this implements.
type Contract interface {
	// GetUserSummary returns a minimal, cross-module-safe view of a user.
	GetUserSummary(ctx context.Context, userID string) (*UserSummary, error)

	// GetPrincipal resolves a user's roles/permissions/scopes for
	// permission checks in other modules.
	GetPrincipal(ctx context.Context, userID string) (*Principal, error)

	// GetUserScopeSet resolves the effective role-scope values a user carries
	// through their active roles, plus whether the user holds a superuser
	// (non-assignable system) role. Used by scope-based data filtering in
	// other modules.
	GetUserScopeSet(ctx context.Context, userID string) (*UserScopeSet, error)

	// GetUserScopeAccess menghitung akses scope user terhadap satu scope type
	// (mis. "gender"). Module pemanggil memakai hasilnya untuk memfilter query
	// resource-nya (IN clause terhadap AllowedCodes, atau tanpa filter saat
	// HasFullAccess).
	GetUserScopeAccess(ctx context.Context, userID, scopeType string) (*UserScopeAccess, error)

	// GetUserScopeAccessIDs sama dengan GetUserScopeAccess tetapi
	// AllowedScopeIDs berisi master scope ID (bukan kode). Dipakai pemanggil
	// yang menyimpan scope_id pada resource untuk auto-tagging & filter.
	GetUserScopeAccessIDs(ctx context.Context, userID, scopeType string) (*UserScopeAccessIDs, error)

	// CanAccessResource mengecek apakah user boleh mengakses sebuah resource
	// yang diklasifikasikan dengan resourceScopeCodes (kode scope master).
	// Resource tanpa scope (list kosong) dianggap publik.
	CanAccessResource(ctx context.Context, userID, scopeType string, resourceScopeCodes []string) (bool, error)

	// CreateAccountWithNIS provisions a brand-new login account (User +
	// Credential + 3 LoginIdentity: NIS primary, email, username) for a
	// caller module that needs to onboard a person who has no existing
	// account yet (e.g. kesantrian's admin-create-santri flow).
	CreateAccountWithNIS(ctx context.Context, in CreateAccountInput) (*CreateAccountResult, error)

	// AddNISLoginIdentity attaches a NIS login identity to an already
	// existing user's local credential (e.g. kesantrian's
	// approve-santri-request flow, where the user already registered
	// normally and is now being recognized as a santri).
	AddNISLoginIdentity(ctx context.Context, userID, nisValue string) error

	// UpdateFullname updates a user's display name on behalf of a caller
	// module that keeps its own profile-like data in sync with identity's
	// (e.g. kesantrian's update-profile flow).
	UpdateFullname(ctx context.Context, userID string, fullname string) error

	// ListActiveUserIDs returns all user IDs with ACTIVE status.
	// Used by notification module for broadcast notifications.
	ListActiveUserIDs(ctx context.Context) ([]string, error)
}

var _ Contract = (*Module)(nil)
