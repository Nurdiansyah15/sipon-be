package ports

import (
	"context"

	"sipon-be/internal/modules/identity"
)

// AccountProvisioner is kesantrian's own port, in kesantrian's own
// vocabulary, for everything it needs from the identity module. It's fine
// to return identity.UserSummary/identity.CreateAccountInput/
// identity.CreateAccountResult directly — those are already contract-
// boundary DTOs, so a second translation layer here would add nothing. See
// docs/architecture/module-boundaries.md's worked "content" module example.
type AccountProvisioner interface {
	GetUserSummary(ctx context.Context, userID string) (*identity.UserSummary, error)
	CreateAccountWithNIS(ctx context.Context, in identity.CreateAccountInput) (*identity.CreateAccountResult, error)
	AddNISLoginIdentity(ctx context.Context, userID, nisValue string) error
	UpdateFullname(ctx context.Context, userID string, fullname string) error
}
