package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/santri_registration/constant"
	regRepo "sipon-be/internal/modules/akademik/domain/santri_registration/repository"
	"sipon-be/internal/shared/kernel"
)

type CancelRegistrationUseCase struct {
	registrationRepo regRepo.SantriRegistrationRepository
}

func NewCancelRegistrationUseCase(registrationRepo regRepo.SantriRegistrationRepository) *CancelRegistrationUseCase {
	return &CancelRegistrationUseCase{registrationRepo: registrationRepo}
}

func (uc *CancelRegistrationUseCase) Execute(ctx context.Context, id string) (*dto.SantriRegistrationResponse, error) {
	registration, err := uc.registrationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeSantriRegistrationNotFound)
	}
	if err := registration.Cancel(); err != nil {
		return nil, application.WrapBadRequestErr(err, constant.CodeSantriRegistrationInvalidStatus)
	}
	if err := uc.registrationRepo.Update(ctx, registration); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	return MapSantriRegistrationToResponse(registration), nil
}
