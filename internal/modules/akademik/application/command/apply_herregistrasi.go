package command

import (
	"context"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	periodRepo "sipon-be/internal/modules/akademik/domain/academic_period/repository"
	regConst "sipon-be/internal/modules/akademik/domain/santri_registration/constant"
	regEntity "sipon-be/internal/modules/akademik/domain/santri_registration/entity"
	regRepo "sipon-be/internal/modules/akademik/domain/santri_registration/repository"
	spRepo "sipon-be/internal/modules/akademik/domain/santri_program/repository"
	"sipon-be/internal/shared/kernel"
)

type ApplyHerregistrasiUseCase struct {
	kesantrianReader  ports.KesantrianReader
	periodRepo        periodRepo.AcademicPeriodRepository
	registrationRepo  regRepo.SantriRegistrationRepository
	santriProgramRepo spRepo.SantriProgramRepository
}

func NewApplyHerregistrasiUseCase(
	kesantrianReader ports.KesantrianReader,
	periodRepo periodRepo.AcademicPeriodRepository,
	registrationRepo regRepo.SantriRegistrationRepository,
	santriProgramRepo spRepo.SantriProgramRepository,
) *ApplyHerregistrasiUseCase {
	return &ApplyHerregistrasiUseCase{
		kesantrianReader:  kesantrianReader,
		periodRepo:        periodRepo,
		registrationRepo:  registrationRepo,
		santriProgramRepo: santriProgramRepo,
	}
}

func (uc *ApplyHerregistrasiUseCase) Execute(ctx context.Context, userID string) (*dto.SantriRegistrationResponse, error) {
	info, err := application.ResolveSantriByUserID(ctx, uc.kesantrianReader, userID)
	if err != nil {
		return nil, err
	}

	period, err := uc.periodRepo.FindOpen(ctx)
	if err != nil {
		if application.IsNotFoundErr(err, application.PeriodNotFoundCode) {
			return nil, kernel.New(application.ErrCodeNotFound)
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	existing, err := uc.registrationRepo.FindBySantriAndPeriod(ctx, info.SantriID, period.ID)
	if err != nil && !application.IsNotFoundErr(err, regConst.CodeSantriRegistrationNotFound) {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	if existing != nil {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	if _, err := uc.santriProgramRepo.FindActiveBySantriID(ctx, info.SantriID); err != nil {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	registration, err := regEntity.NewDraftSantriRegistration(uuid.NewString(), info.SantriID, period.ID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	if err := uc.registrationRepo.Save(ctx, registration); err != nil {
		return nil, application.WrapConflictErr(err, regConst.CodeSantriRegistrationDuplicate)
	}

	resp := MapSantriRegistrationToResponse(registration)
	resp.PeriodName = period.Name
	resp.SantriNIS = info.NIS
	resp.SantriName = info.Fullname
	return resp, nil
}
