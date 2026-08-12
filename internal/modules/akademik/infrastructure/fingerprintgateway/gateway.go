package fingerprintgateway

import (
	"context"
	"time"

	"sipon-be/internal/modules/akademik/application/ports"
	"sipon-be/internal/modules/fingerprint"
)

type Gateway struct {
	contract fingerprint.Contract
}

func New(contract fingerprint.Contract) *Gateway {
	return &Gateway{contract: contract}
}

func (g *Gateway) ListDistinctPinInRange(ctx context.Context, from, to time.Time) ([]ports.FingerprintScanPin, error) {
	results, err := g.contract.ListDistinctPinInRange(ctx, from, to)
	if err != nil {
		return nil, err
	}
	pins := make([]ports.FingerprintScanPin, len(results))
	for i, r := range results {
		pins[i] = ports.FingerprintScanPin{PIN: r.PIN, FirstScanAt: r.FirstScanAt}
	}
	return pins, nil
}
