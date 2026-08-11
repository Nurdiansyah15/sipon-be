package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	periodRepo "sipon-be/internal/modules/akademik/domain/academic_period/repository"
	"sipon-be/internal/modules/akademik/domain/santri_registration/constant"
	regRepo "sipon-be/internal/modules/akademik/domain/santri_registration/repository"
)

type GetSantriRegistrationUseCase struct {
	registrationRepo regRepo.SantriRegistrationRepository
	periodRepo       periodRepo.AcademicPeriodRepository
}

func NewGetSantriRegistrationUseCase(registrationRepo regRepo.SantriRegistrationRepository, periodRepo periodRepo.AcademicPeriodRepository) *GetSantriRegistrationUseCase {
	return &GetSantriRegistrationUseCase{registrationRepo: registrationRepo, periodRepo: periodRepo}
}

func (uc *GetSantriRegistrationUseCase) Execute(ctx context.Context, id string) (*dto.SantriRegistrationResponse, error) {
	registration, err := uc.registrationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeSantriRegistrationNotFound)
	}
	resp := command.MapSantriRegistrationToResponse(registration)
	if period, err := uc.periodRepo.FindByID(ctx, registration.AcademicPeriodID); err == nil {
		resp.PeriodName = period.Name
	}
	return resp, nil
}
