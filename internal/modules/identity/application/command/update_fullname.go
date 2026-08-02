package command

import (
	"context"
	"errors"
	"strings"
	"time"

	"sipon-be/internal/modules/identity/application"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
	"sipon-be/internal/shared/kernel"
)

// UpdateFullnameUseCase backs identity's cross-module Contract — it lets a
// caller module (e.g. kesantrian's update-profile flow) keep a user's
// display name in sync with its own profile data.
type UpdateFullnameUseCase struct {
	userRepo userrepo.UserRepository
}

func NewUpdateFullnameUseCase(userRepo userrepo.UserRepository) *UpdateFullnameUseCase {
	return &UpdateFullnameUseCase{userRepo: userRepo}
}

func (uc *UpdateFullnameUseCase) Execute(ctx context.Context, userID, fullname string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return kernel.New(application.ErrCodeBadRequest)
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeInvalidLoginIdentityValue:
				return kernel.Wrap(application.ErrCodeNotFound, err)
			}
		}
		return kernel.Wrap(application.ErrCodeInternal, err)
	}

	name := fullname
	user.Fullname = &name
	user.UpdatedAt = time.Now()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}

	return nil
}
