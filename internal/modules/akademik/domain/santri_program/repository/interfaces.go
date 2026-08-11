package repository

import (
	"context"

	"sipon-be/internal/modules/akademik/domain/santri_program/entity"
)

type SantriProgramRepository interface {
	Save(ctx context.Context, sp *entity.SantriProgram) error
	FindActiveBySantriID(ctx context.Context, santriID string) (*entity.SantriProgram, error)
	FindBySantriID(ctx context.Context, santriID string) ([]*entity.SantriProgram, error)
	DeactivateAllBySantriID(ctx context.Context, santriID string) error
}
