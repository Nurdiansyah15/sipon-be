package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"
)

type UpdateProfileUseCase struct {
	userRepo domain.UserRepository
}

func NewUpdateProfileUseCase(userRepo domain.UserRepository) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{userRepo: userRepo}
}

func (uc *UpdateProfileUseCase) Execute(ctx context.Context, userID string, req dto.UpdateProfileRequest) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return kernel.Wrap(application.ErrCodeNotFound, err)
	}

	newEmail, err := domain.NewEmail(req.Email)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	if newEmail.String() != user.Email.String() {
		emailExists, err := uc.userRepo.ExistsByLoginIdentity(ctx, domain.LoginIdentifierKindEmail, newEmail.String())
		if err != nil {
			return err
		}
		if emailExists {
			return kernel.New(application.ErrCodeConflict)
		}
	}

	if req.Phone != "" {
		newPhone, err := domain.NewPhoneNumber(req.Phone)
		if err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
			}
			return kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
		}

		currentPhone := ""
		if user.PhoneNumber != nil {
			currentPhone = user.PhoneNumber.String()
		}

		if newPhone.String() != currentPhone {
			phoneExists, err := uc.userRepo.ExistsByLoginIdentity(ctx, domain.LoginIdentifierKindPhone, newPhone.String())
			if err != nil {
				return err
			}
			if phoneExists {
				return kernel.New(application.ErrCodeConflict)
			}
			user.PhoneNumber = &newPhone
		}
	} else {
		user.PhoneNumber = nil
	}

	user.Email = newEmail
	user.Fullname = strPtr(req.Fullname)

	return uc.userRepo.Update(ctx, user)
}
