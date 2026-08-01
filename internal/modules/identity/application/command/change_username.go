package command

import (
	"context"
	"errors"
	"strings"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"
)

type ChangeUsernameUseCase struct {
	userRepo domain.UserRepository
}

func NewChangeUsernameUseCase(userRepo domain.UserRepository) *ChangeUsernameUseCase {
	return &ChangeUsernameUseCase{userRepo: userRepo}
}

func (uc *ChangeUsernameUseCase) Execute(ctx context.Context, userID string, req dto.ChangeUsernameRequest) (*dto.ChangeUsernameResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, kernel.New(application.ErrCodeUnauthorized)
	}

	newUsernameStr := strings.TrimSpace(req.Username)
	newUsername, err := domain.NewUsername(newUsernameStr)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeInvalidLoginIdentityValue:
				return nil, kernel.Wrap(application.ErrCodeNotFound, err)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if user.Username.String() == newUsername.String() {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	existingUser, findErr := uc.userRepo.FindByUsername(ctx, newUsername.String())
	if findErr == nil && existingUser.ID != userID {
		return nil, kernel.New(application.ErrCodeConflict)
	}
	if findErr != nil {
		var ke *kernel.AppError
		if !errors.As(findErr, &ke) || ke.Code != domain.ErrCodeUserNotActive {
			return nil, kernel.Wrap(application.ErrCodeInternal, findErr)
		}
	}

	user.ChangeUsername(newUsername)
	if err := uc.userRepo.UpdateUsername(ctx, user.ID, newUsername.String()); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.ChangeUsernameResponse{
		Message:  "username berhasil diubah",
		Username: newUsername.String(),
	}, nil
}
