package command

import (
	"context"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/program/constant"
	"sipon-be/internal/modules/akademik/domain/program/entity"
	repo "sipon-be/internal/modules/akademik/domain/program/repository"
	"sipon-be/internal/shared/kernel"
)

type CreateProgramUseCase struct {
	programRepo repo.ProgramRepository
}

func NewCreateProgramUseCase(programRepo repo.ProgramRepository) *CreateProgramUseCase {
	return &CreateProgramUseCase{programRepo: programRepo}
}

func (uc *CreateProgramUseCase) Execute(ctx context.Context, req dto.CreateProgramRequest) (*dto.ProgramResponse, error) {
	existing, err := uc.programRepo.FindByCode(ctx, req.Code)
	if err != nil && !application.IsNotFoundErr(err, constant.CodeProgramNotFound) {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	if existing != nil {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	program, err := entity.NewProgram(uuid.NewString(), req.Code, req.Name)
	if err != nil {
		return nil, application.WrapBadRequestErr(err, constant.CodeProgramNotFound)
	}
	if err := uc.programRepo.Save(ctx, program); err != nil {
		return nil, application.WrapConflictErr(err, constant.CodeProgramDuplicate)
	}
	return MapProgramToResponse(program), nil
}

func (uc *CreateProgramUseCase) ExecuteSeed(ctx context.Context, code, name string) error {
	existing, err := uc.programRepo.FindByCode(ctx, code)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}
	program, err := entity.NewProgram(uuid.NewString(), code, name)
	if err != nil {
		return err
	}
	return uc.programRepo.Save(ctx, program)
}

func MapProgramToResponse(p *entity.Program) *dto.ProgramResponse {
	return &dto.ProgramResponse{
		ID:        p.ID,
		Code:      p.Code,
		Name:      p.Name,
		Status:    string(p.Status),
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
