package query

import (
	"context"
	"errors"
	"strings"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
	"sipon-be/internal/shared/kernel"
)

// GetUserSummaryUseCase backs identity's cross-module Contract — it must stay
// minimal (no roles, no timestamps) so the contract never gets pulled along
// when the admin-facing GetUserUseCase's response shape changes.
type GetUserSummaryUseCase struct {
	userRepo userrepo.UserRepository
}

func NewGetUserSummaryUseCase(userRepo userrepo.UserRepository) *GetUserSummaryUseCase {
	return &GetUserSummaryUseCase{userRepo: userRepo}
}

func (uc *GetUserSummaryUseCase) Execute(ctx context.Context, userID string) (*dto.UserSummaryResult, error) {
	user, err := uc.userRepo.FindByID(ctx, strings.TrimSpace(userID))
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeInvalidLoginIdentityValue:
				return nil, kernel.Wrap(application.ErrCodeNotFound, err)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.UserSummaryResult{
		ID:       user.ID,
		Username: user.Username.String(),
		Email:    user.Email.String(),
		Status:   string(user.Status),
	}, nil
}
