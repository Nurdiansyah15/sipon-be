package dokumenasetgateway

import (
	"context"

	dokumenAset "sipon-be/internal/modules/dokumen_aset"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
)

type Gateway struct {
	contract dokumenAset.Contract
}

func New(contract dokumenAset.Contract) *Gateway {
	return &Gateway{contract: contract}
}

func (g *Gateway) GetDownloadURL(ctx context.Context, id string, isAuthenticated bool) (*ports.DokumenAsetDownloadResult, error) {
	result, err := g.contract.GetDownloadURL(ctx, id, isAuthenticated)
	if err != nil {
		return nil, err
	}
	return &ports.DokumenAsetDownloadResult{
		AccessURL: result.AccessURL,
		ExpiresIn: result.ExpiresIn,
	}, nil
}
