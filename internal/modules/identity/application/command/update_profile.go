package command

import (
	"context"
	"errors"
	"time"

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

func (uc *UpdateProfileUseCase) Execute(ctx context.Context, userID string, req dto.UpdateProfileRequest) (*dto.UpdateProfileResponse, error) {
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

	cred := user.FindCredential(domain.CredentialTypeLocal)
	if cred == nil {
		return nil, kernel.New(application.ErrCodeNotFound)
	}

	if req.Fullname != nil {
		user.Fullname = req.Fullname
	}

	if req.Email != nil {
		currentEmailIdentity := cred.FindLoginIdentity(domain.LoginIdentifierKindEmail, user.Email.String())
		if currentEmailIdentity != nil && currentEmailIdentity.IsVerified() {
			return nil, kernel.New(application.ErrCodeConflict)
		}

		newEmail, err := domain.NewEmail(*req.Email)
		if err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
			}
			return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
		}

		existingUser, findErr := uc.userRepo.FindByIdentity(ctx, domain.LoginIdentifierKindEmail, newEmail.String())
		if findErr == nil && existingUser.ID != userID {
			return nil, kernel.New(application.ErrCodeConflict)
		}

		user.Email = newEmail
		if currentEmailIdentity != nil {
			currentEmailIdentity.Value = newEmail.String()
		}
	}

	if req.Phone != nil {
		if user.PhoneNumber != nil {
			currentPhoneIdentity := cred.FindLoginIdentity(domain.LoginIdentifierKindPhone, user.PhoneNumber.String())
			if currentPhoneIdentity != nil && currentPhoneIdentity.IsVerified() {
				return nil, kernel.New(application.ErrCodeConflict)
			}
		}

		newPhone, err := domain.NewPhoneNumber(*req.Phone)
		if err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
			}
			return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
		}

		existingUser, findErr := uc.userRepo.FindByIdentity(ctx, domain.LoginIdentifierKindPhone, newPhone.String())
		if findErr == nil && existingUser.ID != userID {
			return nil, kernel.New(application.ErrCodeConflict)
		}

		user.PhoneNumber = &newPhone
		existingPhoneIdentity := cred.FindLoginIdentityByKind(domain.LoginIdentifierKindPhone)
		if existingPhoneIdentity != nil {
			existingPhoneIdentity.Value = newPhone.String()
		}
	}

	user.UpdatedAt = time.Now()
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.UpdateProfileResponse{Message: "profil berhasil diperbarui"}, nil
}
