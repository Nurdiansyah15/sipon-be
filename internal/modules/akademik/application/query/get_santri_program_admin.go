package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	progConst "sipon-be/internal/modules/akademik/domain/program/constant"
	progRepo "sipon-be/internal/modules/akademik/domain/program/repository"
	spConst "sipon-be/internal/modules/akademik/domain/santri_program/constant"
	spRepo "sipon-be/internal/modules/akademik/domain/santri_program/repository"
	"sipon-be/internal/shared/kernel"
)

// GetSantriProgramAdminUseCase menampilkan program aktif seorang santri untuk
// sisi admin.
type GetSantriProgramAdminUseCase struct {
	santriProgramRepo spRepo.SantriProgramRepository
	programRepo       progRepo.ProgramRepository
}

func NewGetSantriProgramAdminUseCase(
	santriProgramRepo spRepo.SantriProgramRepository,
	programRepo progRepo.ProgramRepository,
) *GetSantriProgramAdminUseCase {
	return &GetSantriProgramAdminUseCase{santriProgramRepo: santriProgramRepo, programRepo: programRepo}
}

func (uc *GetSantriProgramAdminUseCase) Execute(ctx context.Context, santriID string) (*dto.SantriProgramAdminResponse, error) {
	sp, err := uc.santriProgramRepo.FindActiveBySantriID(ctx, santriID)
	if err != nil {
		if application.IsNotFoundErr(err, spConst.CodeSantriProgramNotFound) {
			return nil, kernel.New(application.ErrCodeNotFound)
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	prog, err := uc.programRepo.FindByID(ctx, sp.ProgramID)
	if err != nil {
		if application.IsNotFoundErr(err, progConst.CodeProgramNotFound) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "program tidak ditemukan", err)
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.SantriProgramAdminResponse{
		SantriID:  sp.SantriID,
		ProgramID: prog.ID,
		Program: dto.ProgramBrief{
			ID:   prog.ID,
			Code: prog.Code,
			Name: prog.Name,
		},
		IsActive: sp.IsActive,
	}, nil
}
