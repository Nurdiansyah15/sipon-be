package ports

import (
	"context"

	"sipon-be/internal/modules/identity"
)

// IdentityReader is feedback's own port for everything it needs from the
// identity module — resolving user_id to display names. Returning
// identity.UserSummary directly is fine: it is already a contract-boundary
// DTO (see docs/architecture/module-boundaries.md).
type IdentityReader interface {
	GetUserSummary(ctx context.Context, userID string) (*identity.UserSummary, error)
}
