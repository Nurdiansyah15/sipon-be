package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/program/repository"
	"sipon-be/internal/shared/kernel"
)

type ListProgramsUseCase struct {
	programRepo repository.ProgramRepository
}

func NewListProgramsUseCase(programRepo repository.ProgramRepository) *ListProgramsUseCase {
	return &ListProgramsUseCase{programRepo: programRepo}
}

func (uc *ListProgramsUseCase) Execute(ctx context.Context, q dto.ProgramListQuery) ([]dto.ProgramResponse, *dto.Meta, error) {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.Limit == 0 {
		q.Limit = 20
	}

	result, err := uc.programRepo.List(ctx, repository.ProgramListQuery{
		Status: q.Status,
		Search: q.Search,
		Page:   q.Page,
		Limit:  q.Limit,
	})
	if err != nil {
		return nil, nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := make([]dto.ProgramResponse, len(result.Items))
	for i, p := range result.Items {
		items[i] = *command.MapProgramToResponse(p)
	}
	return items, dto.NewMeta(q.Page, q.Limit, result.Total), nil
}
