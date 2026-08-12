package command

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	progConst "sipon-be/internal/modules/akademik/domain/program/constant"
	progEntity "sipon-be/internal/modules/akademik/domain/program/entity"
	progRepo "sipon-be/internal/modules/akademik/domain/program/repository"
	spConst "sipon-be/internal/modules/akademik/domain/santri_program/constant"
	spEntity "sipon-be/internal/modules/akademik/domain/santri_program/entity"
	spRepo "sipon-be/internal/modules/akademik/domain/santri_program/repository"
	"sipon-be/internal/shared/kernel"
)

// AssignSantriProgramAdminUseCase menetapkan program aktif untuk santri dari
// sisi admin. Operasi deactivate + create dibungkus transactor agar atomic.
type AssignSantriProgramAdminUseCase struct {
	santriProgramRepo spRepo.SantriProgramRepository
	programRepo       progRepo.ProgramRepository
	transactor        ports.Transactor
}

func NewAssignSantriProgramAdminUseCase(
	santriProgramRepo spRepo.SantriProgramRepository,
	programRepo progRepo.ProgramRepository,
	transactor ports.Transactor,
) *AssignSantriProgramAdminUseCase {
	return &AssignSantriProgramAdminUseCase{santriProgramRepo: santriProgramRepo, programRepo: programRepo, transactor: transactor}
}

func (uc *AssignSantriProgramAdminUseCase) Execute(ctx context.Context, santriID, programID string) (*dto.SantriProgramAdminResponse, error) {
	prog, err := uc.programRepo.FindByID(ctx, programID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case progConst.CodeProgramNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "program tidak ditemukan", ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if prog.Status != progConst.ProgramStatusActive {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "program harus berstatus aktif", nil)
	}

	existing, err := uc.santriProgramRepo.FindActiveBySantriID(ctx, santriID)
	if err != nil && !application.IsNotFoundErr(err, spConst.CodeSantriProgramNotFound) {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if existing != nil && existing.ProgramID == programID {
		return mapSantriProgramAdmin(santriID, prog), nil
	}

	err = uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if existing != nil {
			if err := uc.santriProgramRepo.DeactivateAllBySantriID(txCtx, santriID); err != nil {
				return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
			}
		}
		sp, err := spEntity.NewSantriProgram(uuid.NewString(), santriID, programID)
		if err != nil {
			return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
		if err := uc.santriProgramRepo.Save(txCtx, sp); err != nil {
			return application.WrapConflictErr(err, spConst.CodeSantriProgramDuplicate)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return mapSantriProgramAdmin(santriID, prog), nil
}

func mapSantriProgramAdmin(santriID string, prog *progEntity.Program) *dto.SantriProgramAdminResponse {
	return &dto.SantriProgramAdminResponse{
		SantriID:  santriID,
		ProgramID: prog.ID,
		Program: dto.ProgramBrief{
			ID:   prog.ID,
			Code: prog.Code,
			Name: prog.Name,
		},
		IsActive: true,
	}
}
