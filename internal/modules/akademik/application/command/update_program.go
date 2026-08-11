package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/program/constant"
	repo "sipon-be/internal/modules/akademik/domain/program/repository"
	"sipon-be/internal/shared/kernel"
)

type UpdateProgramUseCase struct {
	programRepo repo.ProgramRepository
}

func NewUpdateProgramUseCase(programRepo repo.ProgramRepository) *UpdateProgramUseCase {
	return &UpdateProgramUseCase{programRepo: programRepo}
}

func (uc *UpdateProgramUseCase) Execute(ctx context.Context, id string, req dto.UpdateProgramRequest) (*dto.ProgramResponse, error) {
	program, err := uc.programRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeProgramNotFound)
	}
	if program == nil {
		return nil, kernel.New(application.ErrCodeNotFound)
	}

	if req.Code != nil && *req.Code != "" && *req.Code != program.Code {
		existing, _ := uc.programRepo.FindByCode(ctx, *req.Code)
		if existing != nil && existing.ID != id {
			return nil, kernel.New(application.ErrCodeConflict)
		}
		program.Code = *req.Code
	}

	name := ""
	if req.Name != nil {
		name = *req.Name
	}
	status := ""
	if req.Status != nil {
		status = *req.Status
	}
	if err := program.Update(name, status); err != nil {
		return nil, application.WrapBadRequestErr(err, constant.CodeProgramInvalidStatus)
	}

	if err := uc.programRepo.Update(ctx, program); err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeProgramNotFound)
	}
	return MapProgramToResponse(program), nil
}
