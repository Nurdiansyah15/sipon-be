package repository

import (
	"context"

	"sipon-be/internal/modules/akademik/domain/santri_program/entity"
)

type SantriProgramRepository interface {
	Save(ctx context.Context, sp *entity.SantriProgram) error
	FindActiveBySantriID(ctx context.Context, santriID string) (*entity.SantriProgram, error)
	FindBySantriID(ctx context.Context, santriID string) ([]*entity.SantriProgram, error)
	// ListActiveSantriIDsByProgramID returns all santri IDs with active program
	// membership for the given program.
	ListActiveSantriIDsByProgramID(ctx context.Context, programID string) ([]string, error)
	// ListActive returns all active santri program records (non-deleted).
	ListActive(ctx context.Context) ([]*entity.SantriProgram, error)
	DeactivateAllBySantriID(ctx context.Context, santriID string) error
}
