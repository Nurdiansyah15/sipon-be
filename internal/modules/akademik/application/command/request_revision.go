package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/santri_registration/constant"
	regRepo "sipon-be/internal/modules/akademik/domain/santri_registration/repository"
	"sipon-be/internal/shared/kernel"
)

type RequestRevisionUseCase struct {
	registrationRepo regRepo.SantriRegistrationRepository
}

func NewRequestRevisionUseCase(registrationRepo regRepo.SantriRegistrationRepository) *RequestRevisionUseCase {
	return &RequestRevisionUseCase{registrationRepo: registrationRepo}
}

func (uc *RequestRevisionUseCase) Execute(ctx context.Context, id, notes string) (*dto.SantriRegistrationResponse, error) {
	if notes == "" {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}
	registration, err := uc.registrationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeSantriRegistrationNotFound)
	}
	if err := registration.RequestRevision(notes); err != nil {
		return nil, application.WrapBadRequestErr(err, constant.CodeSantriRegistrationInvalidStatus)
	}
	if err := uc.registrationRepo.Update(ctx, registration); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	return MapSantriRegistrationToResponse(registration), nil
}
