package identitygateway

import (
	"context"

	"sipon-be/internal/modules/identity"
)

// Gateway adapts identity.Contract to kesantrian's own
// ports.AccountProvisioner — the template from
// docs/architecture/module-boundaries.md's "content" module example.
type Gateway struct {
	contract identity.Contract
}

func New(contract identity.Contract) *Gateway {
	return &Gateway{contract: contract}
}

func (g *Gateway) GetUserSummary(ctx context.Context, userID string) (*identity.UserSummary, error) {
	return g.contract.GetUserSummary(ctx, userID)
}

func (g *Gateway) CreateAccountWithNIS(ctx context.Context, in identity.CreateAccountInput) (*identity.CreateAccountResult, error) {
	return g.contract.CreateAccountWithNIS(ctx, in)
}

func (g *Gateway) AddNISLoginIdentity(ctx context.Context, userID, nisValue string) error {
	return g.contract.AddNISLoginIdentity(ctx, userID, nisValue)
}

func (g *Gateway) UpdateFullname(ctx context.Context, userID string, fullname string) error {
	return g.contract.UpdateFullname(ctx, userID, fullname)
}
