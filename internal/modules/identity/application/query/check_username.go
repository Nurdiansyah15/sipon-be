package query

import (
	"context"
	"errors"
	"strings"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
	uservo "sipon-be/internal/modules/identity/domain/user/valueobject"
	"sipon-be/internal/shared/kernel"
)

type CheckUsernameUseCase struct {
	userRepo userrepo.UserRepository
}

func NewCheckUsernameUseCase(userRepo userrepo.UserRepository) *CheckUsernameUseCase {
	return &CheckUsernameUseCase{userRepo: userRepo}
}

func (uc *CheckUsernameUseCase) Execute(ctx context.Context, userID, username string) (*dto.CheckUsernameResponse, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	_, err := uservo.NewUsername(username)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
		}
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	existingUser, err := uc.userRepo.FindByUsername(ctx, username)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
		case userconstant.ErrCodeUserNotFound:
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
