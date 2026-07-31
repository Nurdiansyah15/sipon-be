package query

import (
	"context"
	"errors"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"
)

type CheckUsernameUseCase struct {
	userRepo domain.UserRepository
}

func NewCheckUsernameUseCase(userRepo domain.UserRepository) *CheckUsernameUseCase {
	return &CheckUsernameUseCase{userRepo: userRepo}
}

func (uc *CheckUsernameUseCase) Execute(ctx context.Context, userID, username string) (*dto.CheckUsernameResponse, error) {
	newUsername, err := domain.NewUsername(username)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	existingUser, err := uc.userRepo.FindByUsername(ctx, newUsername.String())
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == domain.ErrCodeUserNotActive {
			return &dto.CheckUsernameResponse{Available: true}, nil
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if existingUser.ID == userID {
		return &dto.CheckUsernameResponse{Available: true}, nil
	}

	return &dto.CheckUsernameResponse{Available: false}, nil
}
