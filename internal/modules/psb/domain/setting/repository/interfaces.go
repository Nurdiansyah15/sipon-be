package repository

import (
	"context"

	"sipon-be/internal/modules/psb/domain/setting/entity"
)

type PsbSettingRepository interface {
	Save(ctx context.Context, setting *entity.PsbSetting) error
	Update(ctx context.Context, setting *entity.PsbSetting) error
	FindByID(ctx context.Context, id string) (*entity.PsbSetting, error)
	FindActive(ctx context.Context) (*entity.PsbSetting, error)
	List(ctx context.Context) ([]*entity.PsbSetting, error)
}
