package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	progRepo "sipon-be/internal/modules/akademik/domain/program/repository"
	spConst "sipon-be/internal/modules/akademik/domain/santri_program/constant"
	spRepo "sipon-be/internal/modules/akademik/domain/santri_program/repository"
	"sipon-be/internal/shared/kernel"
)

type GetMyProgramUseCase struct {
	kesantrianReader  ports.KesantrianReader
	santriProgramRepo spRepo.SantriProgramRepository
	programRepo       progRepo.ProgramRepository
}

func NewGetMyProgramUseCase(
	kesantrianReader ports.KesantrianReader,
	santriProgramRepo spRepo.SantriProgramRepository,
	programRepo progRepo.ProgramRepository,
) *GetMyProgramUseCase {
	return &GetMyProgramUseCase{kesantrianReader: kesantrianReader, santriProgramRepo: santriProgramRepo, programRepo: programRepo}
}

func (uc *GetMyProgramUseCase) Execute(ctx context.Context, userID string) (*dto.MyProgramResponse, error) {
	info, err := application.ResolveSantriByUserID(ctx, uc.kesantrianReader, userID)
	if err != nil {
		return nil, err
	}

	sp, err := uc.santriProgramRepo.FindActiveBySantriID(ctx, info.SantriID)
	if err != nil {
		if application.IsNotFoundErr(err, spConst.CodeSantriProgramNotFound) {
			return nil, kernel.New(application.ErrCodeNotFound)
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	prog, err := uc.programRepo.FindByID(ctx, sp.ProgramID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	started := sp.StartedAt
	return &dto.MyProgramResponse{
		ID:        prog.ID,
		Code:      prog.Code,
		Name:      prog.Name,
		Status:    string(prog.Status),
		StartedAt: &started,
	}, nil
}
