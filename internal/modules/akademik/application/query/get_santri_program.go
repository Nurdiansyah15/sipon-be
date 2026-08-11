package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	progRepo "sipon-be/internal/modules/akademik/domain/program/repository"
	spConst "sipon-be/internal/modules/akademik/domain/santri_program/constant"
	spRepo "sipon-be/internal/modules/akademik/domain/santri_program/repository"
	"sipon-be/internal/shared/kernel"
)

type SantriProgramInfo struct {
	SantriID    string
	ProgramID   string
	ProgramCode string
	ProgramName string
	IsActive    bool
}

type GetSantriProgramUseCase struct {
	santriProgramRepo spRepo.SantriProgramRepository
	programRepo       progRepo.ProgramRepository
}

func NewGetSantriProgramUseCase(santriProgramRepo spRepo.SantriProgramRepository, programRepo progRepo.ProgramRepository) *GetSantriProgramUseCase {
	return &GetSantriProgramUseCase{santriProgramRepo: santriProgramRepo, programRepo: programRepo}
}

func (uc *GetSantriProgramUseCase) Execute(ctx context.Context, santriID string) (*SantriProgramInfo, error) {
	sp, err := uc.santriProgramRepo.FindActiveBySantriID(ctx, santriID)
	if err != nil {
		if application.IsNotFoundErr(err, spConst.CodeSantriProgramNotFound) {
			return nil, kernel.Wrap(application.ErrCodeNotFound, err)
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	info := &SantriProgramInfo{
		SantriID:  sp.SantriID,
		ProgramID: sp.ProgramID,
		IsActive:  sp.IsActive,
	}
	prog, err := uc.programRepo.FindByID(ctx, sp.ProgramID)
	if err == nil {
		info.ProgramCode = prog.Code
		info.ProgramName = prog.Name
	}
	return info, nil
}
