package command

import (
	"context"
	"errors"
	"time"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
	uservo "sipon-be/internal/modules/identity/domain/user/valueobject"
	"sipon-be/internal/shared/kernel"
)

type UpdateProfileUseCase struct {
	userRepo userrepo.UserRepository
}

func NewUpdateProfileUseCase(userRepo userrepo.UserRepository) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{userRepo: userRepo}
}

func (uc *UpdateProfileUseCase) Execute(ctx context.Context, userID string, req dto.UpdateProfileRequest) (*dto.UpdateProfileResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	cred := user.FindCredential(userconstant.CredentialTypeLocal)
	if cred == nil {
		return nil, kernel.New(application.ErrCodeNotFound)
	}

	if req.Fullname != nil {
		user.Fullname = req.Fullname
	}

	if req.Email != nil {
		currentEmailIdentity := cred.FindLoginIdentity(userconstant.LoginIdentifierKindEmail, user.Email.String())
		if currentEmailIdentity != nil && currentEmailIdentity.IsVerified() {
			return nil, kernel.New(application.ErrCodeConflict)
		}

		newEmail, err := uservo.NewEmail(*req.Email)
		if err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
			}
			return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
		}

		existingUser, findErr := uc.userRepo.FindByIdentity(ctx, userconstant.LoginIdentifierKindEmail, newEmail.String())
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
			currentPhoneIdentity := cred.FindLoginIdentity(userconstant.LoginIdentifierKindPhone, user.PhoneNumber.String())
			if currentPhoneIdentity != nil && currentPhoneIdentity.IsVerified() {
				return nil, kernel.New(application.ErrCodeConflict)
			}
		}

		newPhone, err := uservo.NewPhoneNumber(*req.Phone)
		if err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
			}
			return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
		}

		existingUser, findErr := uc.userRepo.FindByIdentity(ctx, userconstant.LoginIdentifierKindPhone, newPhone.String())
		if findErr == nil && existingUser.ID != userID {
			return nil, kernel.New(application.ErrCodeConflict)
		}

		user.PhoneNumber = &newPhone
		existingPhoneIdentity := cred.FindLoginIdentityByKind(userconstant.LoginIdentifierKindPhone)
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
