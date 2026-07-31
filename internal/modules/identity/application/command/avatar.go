package command

import (
	"context"
	"fmt"
	"time"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type AvatarPresignUseCase struct {
	fileUploader application.FileUploader
}

func NewAvatarPresignUseCase(fileUploader application.FileUploader) *AvatarPresignUseCase {
	return &AvatarPresignUseCase{fileUploader: fileUploader}
}

const avatarPresignTTL = 10 * time.Minute

func (uc *AvatarPresignUseCase) Execute(ctx context.Context, req dto.AvatarPresignRequest) (*dto.AvatarPresignResponse, error) {
	objectName := fmt.Sprintf("avatars/%s", uuid.NewString())
	presignURL, key, _, err := uc.fileUploader.RequestUpload(ctx, objectName, req.ContentType, avatarPresignTTL, application.PrivacyPublic)
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
	fileUploader application.FileUploader
}

func NewAvatarConfirmUseCase(
	userRepo domain.UserRepository,
	fileUploader application.FileUploader,
) *AvatarConfirmUseCase {
	return &AvatarConfirmUseCase{
		userRepo:     userRepo,
		fileUploader: fileUploader,
	}
}

func (uc *AvatarConfirmUseCase) Execute(ctx context.Context, userID, key string) (*dto.AvatarConfirmResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	if user.AvatarKey != nil && *user.AvatarKey != "" {
		_ = uc.fileUploader.MarkDeleted(ctx, *user.AvatarKey)
	}

	if err := uc.fileUploader.ConfirmUpload(ctx, key); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	user.AvatarKey = &key

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.AvatarConfirmResponse{AvatarURL: uc.fileUploader.PublicURL(key)}, nil
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

func (uc *AvatarDeleteUseCase) Execute(ctx context.Context, userID string) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return kernel.Wrap(application.ErrCodeNotFound, err)
	}

	if user.AvatarKey == nil || *user.AvatarKey == "" {
		return nil
	}

	if err := uc.fileUploader.DeleteObject(ctx, *user.AvatarKey, application.PrivacyPublic); err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}

	user.AvatarKey = nil

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}

	return nil
}
