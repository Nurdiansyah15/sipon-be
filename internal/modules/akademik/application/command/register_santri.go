package command

import (
	"context"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	periodRepo "sipon-be/internal/modules/akademik/domain/academic_period/repository"
	"sipon-be/internal/modules/akademik/domain/santri_registration/constant"
	"sipon-be/internal/modules/akademik/domain/santri_registration/entity"
	regRepo "sipon-be/internal/modules/akademik/domain/santri_registration/repository"
	"sipon-be/internal/shared/kernel"
)

type RegisterSantriUseCase struct {
	registrationRepo regRepo.SantriRegistrationRepository
	periodRepo       periodRepo.AcademicPeriodRepository
	kesantrianReader ports.KesantrianReader
}

func NewRegisterSantriUseCase(
	registrationRepo regRepo.SantriRegistrationRepository,
	periodRepo periodRepo.AcademicPeriodRepository,
	kesantrianReader ports.KesantrianReader,
) *RegisterSantriUseCase {
	return &RegisterSantriUseCase{registrationRepo: registrationRepo, periodRepo: periodRepo, kesantrianReader: kesantrianReader}
}

func (uc *RegisterSantriUseCase) Execute(ctx context.Context, req dto.CreateSantriRegistrationRequest) (*dto.SantriRegistrationResponse, error) {
	info, err := uc.kesantrianReader.GetSantriByID(ctx, req.SantriID)
	if err != nil {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}
	if info == nil || info.Status != "SANTRI" {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	period, err := uc.periodRepo.FindByID(ctx, req.AcademicPeriodID)
	if err != nil {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}
	if period.Status != "open" {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	existing, err := uc.registrationRepo.FindBySantriAndPeriod(ctx, req.SantriID, req.AcademicPeriodID)
	if err != nil && !application.IsNotFoundErr(err, constant.CodeSantriRegistrationNotFound) {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	if existing != nil {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	registration, err := entity.NewSantriRegistration(uuid.NewString(), req.SantriID, req.AcademicPeriodID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	if err := uc.registrationRepo.Save(ctx, registration); err != nil {
		return nil, application.WrapConflictErr(err, constant.CodeSantriRegistrationDuplicate)
	}

	resp := MapSantriRegistrationToResponse(registration)
	resp.PeriodName = period.Name
	resp.SantriNIS = info.NIS
	resp.SantriName = info.Fullname
	return resp, nil
}

func MapSantriRegistrationToResponse(r *entity.SantriRegistration) *dto.SantriRegistrationResponse {
	return &dto.SantriRegistrationResponse{
		ID:               r.ID,
		SantriID:         r.SantriID,
		AcademicPeriodID: r.AcademicPeriodID,
		Status:           string(r.Status),
		RevisionNotes:    r.RevisionNotes,
		RegisteredAt:     r.RegisteredAt,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}
