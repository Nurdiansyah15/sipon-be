package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	progRepo "sipon-be/internal/modules/akademik/domain/program/repository"
	ptrRepo "sipon-be/internal/modules/akademik/domain/program_transfer_request/repository"
	"sipon-be/internal/shared/kernel"
)

// ListProgramTransferRequestsUseCase menampilkan daftar request transfer
// program untuk admin (filter status, pagination).
type ListProgramTransferRequestsUseCase struct {
	ptrRepo          ptrRepo.ProgramTransferRequestRepository
	programRepo      progRepo.ProgramRepository
	kesantrianReader ports.KesantrianReader
}

func NewListProgramTransferRequestsUseCase(
	ptrRepo ptrRepo.ProgramTransferRequestRepository,
	programRepo progRepo.ProgramRepository,
	kesantrianReader ports.KesantrianReader,
) *ListProgramTransferRequestsUseCase {
	return &ListProgramTransferRequestsUseCase{ptrRepo: ptrRepo, programRepo: programRepo, kesantrianReader: kesantrianReader}
}

func (uc *ListProgramTransferRequestsUseCase) Execute(ctx context.Context, q dto.ProgramTransferRequestListQuery) ([]dto.ProgramTransferRequestResponse, *dto.Meta, error) {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.Limit == 0 {
		q.Limit = 20
	}

	result, err := uc.ptrRepo.List(ctx, ptrRepo.ProgramTransferRequestListQuery{
		Status: q.Status,
		Page:   q.Page,
		Limit:  q.Limit,
	})
	if err != nil {
		return nil, nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := make([]dto.ProgramTransferRequestResponse, 0, len(result.Items))
	for _, req := range result.Items {
		fromProg, _ := uc.programRepo.FindByID(ctx, req.FromProgramID)
		toProg, _ := uc.programRepo.FindByID(ctx, req.ToProgramID)
		var santriName *string
		if info, err := uc.kesantrianReader.GetSantriByID(ctx, req.SantriID); err == nil && info != nil {
			santriName = info.Fullname
		}
		items = append(items, *command.MapProgramTransferRequestToResponse(req, fromProg, toProg, santriName))
	}
	return items, dto.NewMeta(q.Page, q.Limit, result.Total), nil
}
