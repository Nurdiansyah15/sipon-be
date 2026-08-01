package command

import (
	"context"
	"errors"
	"path"
	"strings"
	"time"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

var avatarAllowedContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

var avatarExtByContentType = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

const avatarPresignTTL = 10 * time.Minute

type AvatarPresignUseCase struct {
	fileUploader application.FileUploader
}

func NewAvatarPresignUseCase(fileUploader application.FileUploader) *AvatarPresignUseCase {
	return &AvatarPresignUseCase{fileUploader: fileUploader}
}

func (uc *AvatarPresignUseCase) Execute(ctx context.Context, req dto.AvatarPresignRequest) (*dto.AvatarPresignResponse, error) {
	ct := strings.TrimSpace(strings.ToLower(req.ContentType))
	if !avatarAllowedContentTypes[ct] {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	ext := avatarExtByContentType[ct]
	objectName := path.Join("avatars", uuid.NewString()+ext)

	presignURL, key, _, err := uc.fileUploader.RequestUpload(ctx, objectName, ct, avatarPresignTTL, application.PrivacyPublic)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.AvatarPresignResponse{
		PresignURL: presignURL,
		Key:        key,
		ExpiresIn:  int(avatarPresignTTL.Seconds()),
	}, nil
}

type AvatarConfirmUseCase struct {
	userRepo     domain.UserRepository
	transactor   application.Transactor
	fileUploader application.FileUploader
}

func NewAvatarConfirmUseCase(
	userRepo domain.UserRepository,
	transactor application.Transactor,
	fileUploader application.FileUploader,
) *AvatarConfirmUseCase {
	return &AvatarConfirmUseCase{
		userRepo:     userRepo,
		transactor:   transactor,
		fileUploader: fileUploader,
	}
}

func (uc *AvatarConfirmUseCase) Execute(ctx context.Context, userID, key string) (*dto.AvatarConfirmResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, kernel.New(application.ErrCodeUnauthorized)
	}

	normalizedKey := uc.fileUploader.KeyFromURL(key)
	if normalizedKey == "" {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

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

	oldKey := user.AvatarKey
	user.AvatarKey = &normalizedKey

	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		return uc.userRepo.Update(txCtx, user)
	}); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	_ = uc.fileUploader.ConfirmUpload(ctx, normalizedKey)

	if oldKey != nil && *oldKey != normalizedKey {
		_ = uc.fileUploader.MarkDeleted(ctx, *oldKey)
	}

	return &dto.AvatarConfirmResponse{
		AvatarURL: uc.fileUploader.PublicURL(normalizedKey),
	}, nil
}

type AvatarDeleteUseCase struct {
	userRepo     domain.UserRepository
	fileUploader application.FileUploader
}

func NewAvatarDeleteUseCase(
	userRepo domain.UserRepository,
	fileUploader application.FileUploader,
) *AvatarDeleteUseCase {
	return &AvatarDeleteUseCase{
		userRepo:     userRepo,
		fileUploader: fileUploader,
	}
}

func (uc *AvatarDeleteUseCase) Execute(ctx context.Context, userID string) (*dto.ChangeIdentityResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, kernel.New(application.ErrCodeUnauthorized)
	}

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

	oldKey := user.AvatarKey
	user.AvatarKey = nil

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if oldKey != nil && *oldKey != "" {
		_ = uc.fileUploader.MarkDeleted(ctx, *oldKey)
	}

	return &dto.ChangeIdentityResponse{Message: "avatar berhasil dihapus"}, nil
}
