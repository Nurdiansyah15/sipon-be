package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/program/repository"
	"sipon-be/internal/shared/kernel"
)

type GetProgramUseCase struct {
	programRepo repository.ProgramRepository
}

func NewGetProgramUseCase(programRepo repository.ProgramRepository) *GetProgramUseCase {
	return &GetProgramUseCase{programRepo: programRepo}
}

func (uc *GetProgramUseCase) Execute(ctx context.Context, id string) (*dto.ProgramResponse, error) {
	program, err := uc.programRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, kernel.Code("PROGRAM_NOT_FOUND"))
	}
	return command.MapProgramToResponse(program), nil
}
