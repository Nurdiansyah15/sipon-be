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

func (g *Gateway) GetSantriByID(ctx context.Context, santriID string) (*ports.SantriBasicInfo, error) {
	info, err := g.contract.GetSantriByID(ctx, santriID)
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

func (g *Gateway) ListActiveSantriWithUserID(ctx context.Context) ([]ports.SantriBasicInfo, error) {
	results, err := g.contract.ListActiveSantriWithUserID(ctx)
	if err != nil {
		return nil, err
	}
	infos := make([]ports.SantriBasicInfo, len(results))
	for i, r := range results {
		infos[i] = ports.SantriBasicInfo{
			SantriID: r.SantriID,
			UserID:   r.UserID,
			NIS:      r.NIS,
			Status:   r.Status,
		}
	}
	return infos, nil
}
