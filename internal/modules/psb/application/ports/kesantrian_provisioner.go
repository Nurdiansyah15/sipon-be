package ports

import (
	"context"

	"sipon-be/internal/modules/kesantrian"
)

type KesantrianProvisioner interface {
	CreateSantriFromPendaftaran(ctx context.Context, in kesantrian.CreateSantriFromPendaftaranInput) (*kesantrian.CreateSantriFromPendaftaranResult, error)
}
