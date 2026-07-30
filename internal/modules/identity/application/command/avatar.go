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

func (uc *AvatarPresignUseCase) Execute(ctx context.Context, req dto.AvatarPresignRequest) (*dto.AvatarPresignResponse, error) {
	objectName := fmt.Sprintf("avatars/%s", uuid.NewString())
	presignURL, key, _, err := uc.fileUploader.RequestUpload(ctx, objectName, req.ContentType, 10*time.Minute, application.PrivacyPublic)
	if err != nil {
		return nil, err
	}

	return &dto.AvatarPresignResponse{
		PresignURL: presignURL,
		Key:        key,
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

func (uc *AvatarConfirmUseCase) Execute(ctx context.Context, userID string, req dto.AvatarConfirmRequest) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return kernel.Wrap(application.ErrCodeUserNotFound, err)
	}

	if user.AvatarKey != nil && *user.AvatarKey != "" {
		_ = uc.fileUploader.MarkDeleted(ctx, *user.AvatarKey)
	}

	if err := uc.fileUploader.ConfirmUpload(ctx, req.Key); err != nil {
		return err
	}

	user.AvatarKey = &req.Key

	return uc.userRepo.Update(ctx, user)
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
		return kernel.Wrap(application.ErrCodeUserNotFound, err)
	}

	if user.AvatarKey == nil || *user.AvatarKey == "" {
		return nil
	}

	if err := uc.fileUploader.DeleteObject(ctx, *user.AvatarKey, application.PrivacyPublic); err != nil {
		return err
	}

	user.AvatarKey = nil

	return uc.userRepo.Update(ctx, user)
}
