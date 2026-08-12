package repository

import (
	"context"

	"sipon-be/internal/modules/akademik/domain/program/entity"
)

type ProgramListQuery struct {
	Status *string
	Search *string
	Page   int
	Limit  int
}

type ProgramListResult struct {
	Items []*entity.Program
	Total int64
}

type ProgramRepository interface {
	Save(ctx context.Context, program *entity.Program) error
	Update(ctx context.Context, program *entity.Program) error
	FindByID(ctx context.Context, id string) (*entity.Program, error)
	FindByCode(ctx context.Context, code string) (*entity.Program, error)
	FindByIDs(ctx context.Context, ids []string) ([]*entity.Program, error)
	// ListActiveIDs returns the IDs of all active programs.
	ListActiveIDs(ctx context.Context) ([]string, error)
	List(ctx context.Context, query ProgramListQuery) (*ProgramListResult, error)
}
