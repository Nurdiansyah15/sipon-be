package repository

import (
	"context"

	"sipon-be/internal/modules/psb/domain/dokumen/constant"
	"sipon-be/internal/modules/psb/domain/dokumen/entity"
)

type PendaftarDokumenRepository interface {
	Save(ctx context.Context, d *entity.PendaftarDokumen) error
	Update(ctx context.Context, d *entity.PendaftarDokumen) error
	FindByID(ctx context.Context, id string) (*entity.PendaftarDokumen, error)
	FindByPendaftarID(ctx context.Context, pendaftarID string) ([]*entity.PendaftarDokumen, error)
	FindByPendaftarIDAndStage(ctx context.Context, pendaftarID string, stage constant.DokumenStage) ([]*entity.PendaftarDokumen, error)
	HardDeleteByPendaftarID(ctx context.Context, pendaftarID string) (int64, error)
}
