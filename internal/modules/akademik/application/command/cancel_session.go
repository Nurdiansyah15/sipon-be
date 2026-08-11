package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/activity_session/constant"
	sesRepo "sipon-be/internal/modules/akademik/domain/activity_session/repository"
	"sipon-be/internal/shared/kernel"
)

type CancelSessionUseCase struct {
	sessionRepo sesRepo.ActivitySessionRepository
}

func NewCancelSessionUseCase(sessionRepo sesRepo.ActivitySessionRepository) *CancelSessionUseCase {
	return &CancelSessionUseCase{sessionRepo: sessionRepo}
}

func (uc *CancelSessionUseCase) Execute(ctx context.Context, id string) (*dto.ActivitySessionResponse, error) {
	session, err := uc.sessionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeActivitySessionNotFound)
	}
	if err := session.Cancel(); err != nil {
		return nil, application.WrapBadRequestErr(err, constant.CodeActivitySessionInvalidStatus)
	}
	if err := uc.sessionRepo.Update(ctx, session); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	return MapSessionToResponse(session), nil
}
