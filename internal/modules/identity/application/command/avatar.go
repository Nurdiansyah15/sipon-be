package command

import (
	"context"
	"errors"
	"log/slog"
	"path"
	"strings"
	"time"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	ports "sipon-be/internal/modules/identity/application/ports"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
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
	fileUploader ports.FileUploader
}

func NewAvatarPresignUseCase(fileUploader ports.FileUploader) *AvatarPresignUseCase {
	return &AvatarPresignUseCase{fileUploader: fileUploader}
}

func (uc *AvatarPresignUseCase) Execute(ctx context.Context, req dto.AvatarPresignRequest) (*dto.AvatarPresignResponse, error) {
	ct := strings.TrimSpace(strings.ToLower(req.ContentType))
	if !avatarAllowedContentTypes[ct] {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	ext := avatarExtByContentType[ct]
	objectName := path.Join("pending", "avatars", uuid.NewString()+ext)

	presignURL, key, _, err := uc.fileUploader.RequestUpload(ctx, objectName, ct, avatarPresignTTL, ports.PrivacyPublic)
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
	userRepo     userrepo.UserRepository
	transactor   ports.Transactor
	fileUploader ports.FileUploader
}

func NewAvatarConfirmUseCase(
	userRepo userrepo.UserRepository,
	transactor ports.Transactor,
	fileUploader ports.FileUploader,
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
	if normalizedKey == "" || !strings.HasPrefix(normalizedKey, "pending/") {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
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

	if err := uc.fileUploader.ConfirmUpload(ctx, normalizedKey); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	oldKey := user.AvatarKey
	finalKey := strings.TrimPrefix(normalizedKey, "pending/")

	if err := uc.fileUploader.PromoteUpload(ctx, normalizedKey, finalKey, ports.PrivacyPublic); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	user.AvatarKey = &finalKey
	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		return uc.userRepo.Update(txCtx, user)
	}); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if oldKey != nil && *oldKey != finalKey {
		if err := uc.fileUploader.MarkDeleted(ctx, *oldKey); err != nil {
			slog.Warn("identity: best-effort hapus avatar lama gagal", "key", *oldKey, "error", err)
		}
	}

	return &dto.AvatarConfirmResponse{
		AvatarURL: uc.fileUploader.PublicURL(finalKey),
	}, nil
}

type AvatarDeleteUseCase struct {
	userRepo     userrepo.UserRepository
	fileUploader ports.FileUploader
}

func NewAvatarDeleteUseCase(
	userRepo userrepo.UserRepository,
	fileUploader ports.FileUploader,
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
			case userconstant.ErrCodeInvalidLoginIdentityValue:
				return nil, kernel.Wrap(application.ErrCodeNotFound, err)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	oldKey := user.AvatarKey

	if oldKey != nil && *oldKey != "" {
		if err := uc.fileUploader.MarkDeleted(ctx, *oldKey); err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
	}

	user.AvatarKey = nil
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.ChangeIdentityResponse{Message: "avatar berhasil dihapus"}, nil
}
