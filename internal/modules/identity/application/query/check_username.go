package query

import (
	"context"
	"errors"
	"strings"

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
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	_, err := domain.NewUsername(username)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	existingUser, err := uc.userRepo.FindByUsername(ctx, username)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeUserNotActive:
				return &dto.CheckUsernameResponse{Available: true}, nil
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if existingUser.ID == userID {
		return &dto.CheckUsernameResponse{Available: true}, nil
	}

	return &dto.CheckUsernameResponse{Available: false}, nil
}
