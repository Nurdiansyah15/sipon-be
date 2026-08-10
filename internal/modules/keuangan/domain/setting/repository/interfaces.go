package repository

import (
	"context"

	"sipon-be/internal/modules/keuangan/domain/setting/entity"
)

type KeuanganSettingRepository interface {
	Find(ctx context.Context) (*entity.KeuanganSetting, error)
	Update(ctx context.Context, setting *entity.KeuanganSetting) error
}
