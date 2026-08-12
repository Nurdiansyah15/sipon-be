package command

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	progConst "sipon-be/internal/modules/akademik/domain/program/constant"
	progRepo "sipon-be/internal/modules/akademik/domain/program/repository"
	ptrConst "sipon-be/internal/modules/akademik/domain/program_transfer_request/constant"
	ptrEntity "sipon-be/internal/modules/akademik/domain/program_transfer_request/entity"
	ptrRepo "sipon-be/internal/modules/akademik/domain/program_transfer_request/repository"
	spConst "sipon-be/internal/modules/akademik/domain/santri_program/constant"
	spRepo "sipon-be/internal/modules/akademik/domain/santri_program/repository"
	"sipon-be/internal/shared/kernel"
)

// RequestProgramTransferUseCase membuat permintaan pindah program dari santri.
type RequestProgramTransferUseCase struct {
	ptrRepo           ptrRepo.ProgramTransferRequestRepository
	santriProgramRepo spRepo.SantriProgramRepository
	programRepo       progRepo.ProgramRepository
	kesantrianReader  ports.KesantrianReader
}

func NewRequestProgramTransferUseCase(
	ptrRepo ptrRepo.ProgramTransferRequestRepository,
	santriProgramRepo spRepo.SantriProgramRepository,
	programRepo progRepo.ProgramRepository,
	kesantrianReader ports.KesantrianReader,
) *RequestProgramTransferUseCase {
	return &RequestProgramTransferUseCase{
		ptrRepo:           ptrRepo,
		santriProgramRepo: santriProgramRepo,
		programRepo:       programRepo,
		kesantrianReader:  kesantrianReader,
	}
}

func (uc *RequestProgramTransferUseCase) Execute(ctx context.Context, userID string, req dto.RequestProgramTransferRequest) (*dto.ProgramTransferRequestResponse, error) {
	info, err := application.ResolveSantriByUserID(ctx, uc.kesantrianReader, userID)
	if err != nil {
		return nil, err
	}
	if info == nil || info.Status != "SANTRI" {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "santri tidak aktif", nil)
	}

	toProg, err := uc.programRepo.FindByID(ctx, req.ToProgramID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case progConst.CodeProgramNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "program tujuan tidak ditemukan", ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if toProg.Status != progConst.ProgramStatusActive {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "program tujuan harus berstatus aktif", nil)
	}

	current, err := uc.santriProgramRepo.FindActiveBySantriID(ctx, info.SantriID)
	if err != nil {
		if application.IsNotFoundErr(err, spConst.CodeSantriProgramNotFound) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "santri belum memiliki program aktif", nil)
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if current.ProgramID == req.ToProgramID {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "santri sudah berada di program tersebut", nil)
	}

	pending, err := uc.ptrRepo.FindPendingBySantriID(ctx, info.SantriID)
	if err != nil && !application.IsNotFoundErr(err, ptrConst.CodeProgramTransferRequestNotFound) {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if pending != nil {
		return nil, kernel.WrapMsg(application.ErrCodeConflict, "masih ada permintaan pindah program yang belum diproses", nil)
	}

	transfer, err := ptrEntity.NewProgramTransferRequest(uuid.NewString(), info.SantriID, current.ProgramID, req.ToProgramID, req.Notes)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == ptrConst.CodeProgramTransferRequestSameProgram {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "program tujuan sama dengan program saat ini", ke)
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if err := uc.ptrRepo.Save(ctx, transfer); err != nil {
		return nil, application.WrapConflictErr(err, ptrConst.CodeProgramTransferRequestDuplicate)
	}

	fromProg, _ := uc.programRepo.FindByID(ctx, current.ProgramID)
	return MapProgramTransferRequestToResponse(transfer, fromProg, toProg, info.Fullname), nil
}
