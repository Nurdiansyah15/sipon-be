package identitygateway

import (
	"context"

	"sipon-be/internal/modules/feedback/application/ports"
	"sipon-be/internal/modules/identity"
)

// Gateway adapts identity.Contract to feedback's own ports.IdentityReader —
// the template from docs/architecture/module-boundaries.md.
type Gateway struct {
	contract identity.Contract
}

func New(contract identity.Contract) *Gateway {
	return &Gateway{contract: contract}
}

func (g *Gateway) GetUserSummary(ctx context.Context, userID string) (*identity.UserSummary, error) {
	return g.contract.GetUserSummary(ctx, userID)
}

var _ ports.IdentityReader = (*Gateway)(nil)
