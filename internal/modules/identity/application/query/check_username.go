package query

import (
	"context"

	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
)

type CheckUsernameUseCase struct {
	userRepo domain.UserRepository
}

func NewCheckUsernameUseCase(userRepo domain.UserRepository) *CheckUsernameUseCase {
	return &CheckUsernameUseCase{userRepo: userRepo}
}

func (uc *CheckUsernameUseCase) Execute(ctx context.Context, username string) (*dto.CheckUsernameResponse, error) {
	_, err := domain.NewUsername(username)
	if err != nil {
		return &dto.CheckUsernameResponse{Available: false}, nil
	}

	exists, err := uc.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	return &dto.CheckUsernameResponse{Available: !exists}, nil
}
