package command

import (
	"context"
	"errors"

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
	newUsername, err := domain.NewUsername(req.Username)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	exists, err := uc.userRepo.ExistsByUsername(ctx, newUsername.String())
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	if exists {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	user.ChangeUsername(newUsername)

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.ChangeUsernameResponse{
		Message:  "username berhasil diubah",
		Username: newUsername.String(),
	}, nil
}
