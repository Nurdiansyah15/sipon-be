package command

import (
	"context"

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

func (uc *ChangeUsernameUseCase) Execute(ctx context.Context, userID string, req dto.ChangeUsernameRequest) error {
	newUsername, err := domain.NewUsername(req.Username)
	if err != nil {
		return err
	}

	exists, err := uc.userRepo.ExistsByUsername(ctx, newUsername.String())
	if err != nil {
		return err
	}
	if exists {
		return kernel.New(application.ErrCodeDuplicateUsername)
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return kernel.Wrap(application.ErrCodeUserNotFound, err)
	}

	user.ChangeUsername(newUsername)

	return uc.userRepo.Update(ctx, user)
}
