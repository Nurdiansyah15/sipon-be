package command

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application"
	progConst "sipon-be/internal/modules/akademik/domain/program/constant"
	progRepo "sipon-be/internal/modules/akademik/domain/program/repository"
	spConst "sipon-be/internal/modules/akademik/domain/santri_program/constant"
	spEntity "sipon-be/internal/modules/akademik/domain/santri_program/entity"
	spRepo "sipon-be/internal/modules/akademik/domain/santri_program/repository"
	"sipon-be/internal/shared/kernel"
)

type AssignSantriProgramUseCase struct {
	santriProgramRepo spRepo.SantriProgramRepository
	programRepo       progRepo.ProgramRepository
}

func NewAssignSantriProgramUseCase(santriProgramRepo spRepo.SantriProgramRepository, programRepo progRepo.ProgramRepository) *AssignSantriProgramUseCase {
	return &AssignSantriProgramUseCase{santriProgramRepo: santriProgramRepo, programRepo: programRepo}
}

// Execute menetapkan program aktif untuk seorang santri. Program lama yang
// aktif otomatis di-deactivate terlebih dahulu (hanya satu program aktif).
func (uc *AssignSantriProgramUseCase) Execute(ctx context.Context, santriID, programID string) error {
	prog, err := uc.programRepo.FindByID(ctx, programID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case progConst.CodeProgramNotFound:
				return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "program tidak ditemukan", ke)
			}
		}
		return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if prog.Status != progConst.ProgramStatusActive {
		return kernel.WrapMsg(application.ErrCodeBadRequest, "program harus berstatus aktif", nil)
	}

	existing, err := uc.santriProgramRepo.FindActiveBySantriID(ctx, santriID)
	if err != nil && !application.IsNotFoundErr(err, spConst.CodeSantriProgramNotFound) {
		return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if existing != nil && existing.ProgramID == programID {
		return nil
	}

	if existing != nil {
		if err := uc.santriProgramRepo.DeactivateAllBySantriID(ctx, santriID); err != nil {
			return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
	}

	sp, err := spEntity.NewSantriProgram(uuid.NewString(), santriID, programID)
	if err != nil {
		return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if err := uc.santriProgramRepo.Save(ctx, sp); err != nil {
		return application.WrapConflictErr(err, spConst.CodeSantriProgramDuplicate)
	}
	return nil
}
