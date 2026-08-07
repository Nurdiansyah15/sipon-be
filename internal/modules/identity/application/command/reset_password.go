package command

import (
	"context"
	"errors"
	"time"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	ports "sipon-be/internal/modules/identity/application/ports"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
	uservo "sipon-be/internal/modules/identity/domain/user/valueobject"
	verificationconstant "sipon-be/internal/modules/identity/domain/verification/constant"
	verificationrepo "sipon-be/internal/modules/identity/domain/verification/repository"
	"sipon-be/internal/shared/kernel"
)

type ResetPasswordUseCase struct {
	userRepo  userrepo.UserRepository
	verifRepo verificationrepo.VerificationRepository
	hasher    ports.PasswordHasher
}

func NewResetPasswordUseCase(
	userRepo userrepo.UserRepository,
	verifRepo verificationrepo.VerificationRepository,
	hasher ports.PasswordHasher,
) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{
		userRepo:  userRepo,
		verifRepo: verifRepo,
		hasher:    hasher,
	}
}

func (uc *ResetPasswordUseCase) Execute(ctx context.Context, req dto.ResetPasswordRequest) (*dto.ResetPasswordResponse, error) {
	user, err := uc.userRepo.FindByIdentity(ctx, userconstant.LoginIdentifierKindEmail, req.Email)
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

	verifCode, err := uc.verifRepo.FindLatestByUserAndPurpose(ctx, user.ID, verificationconstant.PurposeResetPassword)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case verificationconstant.ErrCodeVerificationCodeNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if err := verifCode.Verify(req.Token); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case verificationconstant.ErrCodeVerificationCodeExpired, verificationconstant.ErrCodeVerificationCodeUsed, verificationconstant.ErrCodeVerificationCodeMismatch:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	plainPw, err := uservo.NewPlainPassword(req.Password)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodePlainPasswordEmpty, userconstant.ErrCodePlainPasswordTooShort, userconstant.ErrCodePlainPasswordNoUppercase, userconstant.ErrCodePlainPasswordNoDigit:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	localCred := user.FindCredential(userconstant.CredentialTypeLocal)
	if localCred == nil {
		return nil, kernel.WrapMsg(application.ErrCodeForbidden, "Kredensial lokal tidak ditemukan", nil)
	}

	hashedStr, err := uc.hasher.Hash(plainPw.String())
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal mengenkripsi kata sandi", err)
	}

	hashedPw, err := uservo.NewHashedPassword(hashedStr)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal memvalidasi hash kata sandi", err)
	}

	now := time.Now()
	localCred.SecretHash = &hashedPw
	localCred.LastChangedAt = &now
	localCred.UpdatedAt = now
	user.UpdatedAt = now
	user.ResetFailedAttempts()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal memperbarui kata sandi", err)
	}

	if err := uc.verifRepo.Update(ctx, verifCode); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal memperbarui kode verifikasi", err)
	}

	return &dto.ResetPasswordResponse{Message: "password berhasil direset"}, nil
}
