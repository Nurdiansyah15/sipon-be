package query

import (
	"context"
	"errors"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
	"sipon-be/internal/shared/kernel"
)

type GetLinkedAccountsUseCase struct {
	userRepo userrepo.UserRepository
}

func NewGetLinkedAccountsUseCase(userRepo userrepo.UserRepository) *GetLinkedAccountsUseCase {
	return &GetLinkedAccountsUseCase{userRepo: userRepo}
}

func (uc *GetLinkedAccountsUseCase) Execute(ctx context.Context, userID string) (*dto.LinkedAccountsResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	google := user.FindCredential(userconstant.CredentialTypeGoogle)
	linked := google != nil && google.DeletedAt == nil

	var email *string
	if linked {
		v := user.Email.String()
		email = &v
	}

	return &dto.LinkedAccountsResponse{
		Google: dto.GoogleLinkedAccount{
			Linked:    linked,
			Email:     email,
			CanUnlink: user.HasLocalPassword(),
		},
	}, nil
}
