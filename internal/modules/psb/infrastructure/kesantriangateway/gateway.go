package kesantriangateway

import (
	"context"

	"sipon-be/internal/modules/kesantrian"
)

type Gateway struct {
	contract kesantrian.Contract
}

func New(contract kesantrian.Contract) *Gateway {
	return &Gateway{contract: contract}
}

func (g *Gateway) CreateSantriFromPendaftaran(ctx context.Context, in kesantrian.CreateSantriFromPendaftaranInput) (*kesantrian.CreateSantriFromPendaftaranResult, error) {
	return g.contract.CreateSantriFromPendaftaran(ctx, in)
}

func (g *Gateway) GetSantriByUserID(ctx context.Context, userID string) (*kesantrian.SantriBasicInfo, error) {
	return g.contract.GetSantriByUserID(ctx, userID)
}
