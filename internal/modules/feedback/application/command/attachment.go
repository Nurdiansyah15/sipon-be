package command

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/feedback/application"
	"sipon-be/internal/modules/feedback/application/dto"
	ports "sipon-be/internal/modules/feedback/application/ports"
	acconstant "sipon-be/internal/modules/feedback/domain/attachment/constant"
	aentity "sipon-be/internal/modules/feedback/domain/attachment/entity"
	arepo "sipon-be/internal/modules/feedback/domain/attachment/repository"
	fconstant "sipon-be/internal/modules/feedback/domain/feedback/constant"
	frepo "sipon-be/internal/modules/feedback/domain/feedback/repository"
	"sipon-be/internal/shared/kernel"
)

type AttachmentUseCase struct {
	attachmentRepo arepo.AttachmentRepository
	feedbackRepo   frepo.FeedbackRepository
	fileUploader   ports.FileUploader
	transactor     ports.Transactor
}

func NewAttachmentUseCase(
	attachmentRepo arepo.AttachmentRepository,
	feedbackRepo frepo.FeedbackRepository,
	fileUploader ports.FileUploader,
	transactor ports.Transactor,
) *AttachmentUseCase {
	return &AttachmentUseCase{
		attachmentRepo: attachmentRepo,
		feedbackRepo:   feedbackRepo,
		fileUploader:   fileUploader,
		transactor:     transactor,
	}
}

func (uc *AttachmentUseCase) Presign(ctx context.Context, userID, feedbackID string, req dto.AttachmentPresignRequest) (*dto.AttachmentPresignResponse, error) {
	if !acconstant.AllowedContentTypes[req.ContentType] {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}

	f, err := uc.feedbackRepo.FindByID(ctx, feedbackID)
	if err != nil {
		return nil, application.WrapRepoErr(err, fconstant.CodeFeedbackNotFound)
	}
	if f.UserID != userID {
		return nil, kernel.New(application.ErrCodeForbidden)
	}
	if f.IsTakedown {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	count, err := uc.attachmentRepo.CountByFeedbackID(ctx, feedbackID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	if count >= acconstant.MaxAttachmentsPerFeedback {
		return nil, kernel.New(acconstant.CodeAttachmentLimitExceeded)
	}

	ext := filepath.Ext(req.Filename)
	objectName := fmt.Sprintf("pending/feedback/%s/%s%s", feedbackID, uuid.NewString(), ext)
	expiry := 15 * time.Minute

	presignURL, key, publicURL, err := uc.fileUploader.RequestUpload(ctx, objectName, req.ContentType, expiry, ports.PrivacyPrivate)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.AttachmentPresignResponse{
		PresignURL: presignURL,
		Key:        key,
		PublicURL:  publicURL,
	}, nil
}

func (uc *AttachmentUseCase) Confirm(ctx context.Context, userID, feedbackID string, req dto.AttachmentConfirmRequest) (*dto.AttachmentResponse, error) {
	f, err := uc.feedbackRepo.FindByID(ctx, feedbackID)
	if err != nil {
		return nil, application.WrapRepoErr(err, fconstant.CodeFeedbackNotFound)
	}
	if f.UserID != userID {
		return nil, kernel.New(application.ErrCodeForbidden)
	}
	if f.IsTakedown {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	stagingKey := req.Key
	if !strings.HasPrefix(stagingKey, "pending/feedback/"+feedbackID+"/") {
		slog.Warn("feedback: attachment confirm skip — key bukan staging prefix feedback", "key", stagingKey)
		return nil, kernel.New(application.ErrCodeBadRequest)
	}

	if err := uc.fileUploader.ConfirmUpload(ctx, stagingKey); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	finalKey := strings.TrimPrefix(stagingKey, "pending/")
	if err := uc.fileUploader.PromoteUpload(ctx, stagingKey, finalKey, ports.PrivacyPrivate); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	var result *dto.AttachmentResponse
	err = uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		count, err := uc.attachmentRepo.CountByFeedbackID(txCtx, feedbackID)
		if err != nil {
			return err
		}
		if count >= acconstant.MaxAttachmentsPerFeedback {
			return kernel.New(acconstant.CodeAttachmentLimitExceeded)
		}

		sortOrder, err := uc.attachmentRepo.MaxSortOrder(txCtx, feedbackID)
		if err != nil {
			return err
		}

		var of, mt *string
		var sz *int64
		if req.OriginalFilename != "" {
			of = &req.OriginalFilename
		}
		if req.MimeType != "" {
			mt = &req.MimeType
		}
		if req.Size > 0 {
			s := req.Size
			sz = &s
		}

		a, err := aentity.NewAttachment(uuid.NewString(), feedbackID, finalKey, of, mt, sz, sortOrder+1)
		if err != nil {
			return err
		}
		if err := uc.attachmentRepo.Save(txCtx, a); err != nil {
			return err
		}

		result = &dto.AttachmentResponse{
			ID:               a.ID,
			Key:              a.Key,
			OriginalFilename: a.OriginalFilename,
			MimeType:         a.MimeType,
			Size:             a.Size,
			CreatedAt:        a.CreatedAt,
		}
		return nil
	})
	if err != nil {
		_ = uc.fileUploader.DeleteObject(ctx, finalKey, ports.PrivacyPrivate)
		if application.IsDomainErr(err, acconstant.CodeAttachmentLimitExceeded) {
			return nil, kernel.New(acconstant.CodeAttachmentLimitExceeded)
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return result, nil
}

func (uc *AttachmentUseCase) Delete(ctx context.Context, userID, feedbackID, attachmentID string) (*dto.MessageResponse, error) {
	f, err := uc.feedbackRepo.FindByID(ctx, feedbackID)
	if err != nil {
		return nil, application.WrapRepoErr(err, fconstant.CodeFeedbackNotFound)
	}
	if f.UserID != userID {
		return nil, kernel.New(application.ErrCodeForbidden)
	}

	a, err := uc.attachmentRepo.FindByID(ctx, attachmentID)
	if err != nil {
		return nil, application.WrapRepoErr(err, acconstant.CodeAttachmentNotFound)
	}
	if a.FeedbackID != feedbackID {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}

	if err := uc.attachmentRepo.SoftDelete(ctx, a.ID); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.fileUploader.DeleteObject(ctx, a.Key, ports.PrivacyPrivate); err != nil {
		slog.Warn("feedback: best-effort hapus objek attachment gagal", "key", a.Key, "error", err)
	}

	return &dto.MessageResponse{Message: "attachment berhasil dihapus"}, nil
}
