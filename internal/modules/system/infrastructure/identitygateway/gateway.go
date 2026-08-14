package identitygateway

import (
	"context"

	"sipon-be/internal/modules/identity"
)

// Gateway mengimplementasikan ports.IdentityReader milik module system dengan
// mendelegasikan ke identity.Contract. Lihat docs/architecture/module-boundaries.md.
type Gateway struct {
	contract identity.Contract
}

func New(contract identity.Contract) *Gateway {
	return &Gateway{contract: contract}
}

func (g *Gateway) GetUserScopeSet(ctx context.Context, userID string) (*identity.UserScopeSet, error) {
	return g.contract.GetUserScopeSet(ctx, userID)
}
