package identity

import (
	"context"
)

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
}

var _ Contract = (*Module)(nil)
