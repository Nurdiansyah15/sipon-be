package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/santri_registration/constant"
	regRepo "sipon-be/internal/modules/akademik/domain/santri_registration/repository"
	"sipon-be/internal/shared/kernel"
)

type CompleteRegistrationUseCase struct {
	registrationRepo regRepo.SantriRegistrationRepository
}

func NewCompleteRegistrationUseCase(registrationRepo regRepo.SantriRegistrationRepository) *CompleteRegistrationUseCase {
	return &CompleteRegistrationUseCase{registrationRepo: registrationRepo}
}

func (uc *CompleteRegistrationUseCase) Execute(ctx context.Context, id string) (*dto.SantriRegistrationResponse, error) {
	registration, err := uc.registrationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeSantriRegistrationNotFound)
	}
	if err := registration.Complete(); err != nil {
		return nil, application.WrapBadRequestErr(err, constant.CodeSantriRegistrationInvalidStatus)
	}
	if err := uc.registrationRepo.Update(ctx, registration); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	return MapSantriRegistrationToResponse(registration), nil
}
