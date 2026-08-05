package kesantriangateway

import (
	"context"

	"sipon-be/internal/modules/kesantrian"
	"sipon-be/internal/modules/keuangan/application/ports"
)

type Gateway struct {
	contract kesantrian.Contract
}

func New(contract kesantrian.Contract) *Gateway {
	return &Gateway{contract: contract}
}

func (g *Gateway) ListActiveSantriIDs(ctx context.Context) ([]string, error) {
	return g.contract.ListActiveSantriIDs(ctx)
}

func (g *Gateway) GetSantriByUserID(ctx context.Context, userID string) (*ports.SantriBasicInfo, error) {
	info, err := g.contract.GetSantriByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &ports.SantriBasicInfo{
		SantriID: info.SantriID,
		UserID:   info.UserID,
		NIS:      info.NIS,
		Status:   info.Status,
	}, nil
}
