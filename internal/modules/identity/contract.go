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

type ScopeInfo struct {
	ScopeType string
	ScopeID   *string
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
}

var _ Contract = (*Module)(nil)
