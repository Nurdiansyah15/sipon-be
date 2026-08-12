package kesantriangateway

import (
	"context"

	"sipon-be/internal/modules/akademik/application/ports"
	"sipon-be/internal/modules/kesantrian"
)

type Gateway struct {
	contract kesantrian.Contract
}

func New(contract kesantrian.Contract) *Gateway {
	return &Gateway{contract: contract}
}

func (g *Gateway) GetSantriByID(ctx context.Context, santriID string) (*ports.SantriBasicInfo, error) {
	info, err := g.contract.GetSantriByID(ctx, santriID)
	if err != nil {
		return nil, err
	}
	return ports.FromKesantrian(info), nil
}

func (g *Gateway) GetSantriByUserID(ctx context.Context, userID string) (*ports.SantriBasicInfo, error) {
	info, err := g.contract.GetSantriByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return ports.FromKesantrian(info), nil
}

func (g *Gateway) GetSantriByNIS(ctx context.Context, nis string) (*ports.SantriBasicInfo, error) {
	info, err := g.contract.GetSantriByNIS(ctx, nis)
	if err != nil {
		return nil, err
	}
	return ports.FromKesantrian(info), nil
}

func (g *Gateway) ListActiveSantriWithUserID(ctx context.Context) ([]ports.SantriBasicInfo, error) {
	results, err := g.contract.ListActiveSantriWithUserID(ctx)
	if err != nil {
		return nil, err
	}
	infos := make([]ports.SantriBasicInfo, len(results))
	for i, r := range results {
		infos[i] = *ports.FromKesantrian(&r)
	}
	return infos, nil
}
