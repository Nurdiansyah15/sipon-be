package repository

import (
	"context"

	"sipon-be/internal/modules/psb/domain/pendaftar/entity"
)

type PendaftarListQuery struct {
	PsbSettingID string
	Status       *string
	Page         int
	Limit        int
}

type PendaftarListResult struct {
	Items []*entity.Pendaftar
	Total int64
}

type PendaftarRepository interface {
	Save(ctx context.Context, p *entity.Pendaftar) error
	Update(ctx context.Context, p *entity.Pendaftar) error
	FindByID(ctx context.Context, id string) (*entity.Pendaftar, error)
	FindByUserIDAndSetting(ctx context.Context, userID, psbSettingID string) (*entity.Pendaftar, error)
	CountBySettingAndProgram(ctx context.Context, psbSettingID, program string) (int64, error)
	List(ctx context.Context, query PendaftarListQuery) (*PendaftarListResult, error)
	HardDeleteBySettingID(ctx context.Context, psbSettingID string) (int64, error)
}
