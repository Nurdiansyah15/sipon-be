package repository

import (
	"context"

	"sipon-be/internal/modules/akademik/domain/setting/entity"
)

type AkademikSettingRepository interface {
	Find(ctx context.Context) (*entity.AkademikSetting, error)
	Update(ctx context.Context, setting *entity.AkademikSetting) error
}
