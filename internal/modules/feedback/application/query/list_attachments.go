package query

import (
	"context"
	"log/slog"

	"sipon-be/internal/modules/feedback/application"
	"sipon-be/internal/modules/feedback/application/dto"
	ports "sipon-be/internal/modules/feedback/application/ports"
	arepo "sipon-be/internal/modules/feedback/domain/attachment/repository"
	fconstant "sipon-be/internal/modules/feedback/domain/feedback/constant"
	frepo "sipon-be/internal/modules/feedback/domain/feedback/repository"
	"sipon-be/internal/shared/kernel"
)

type ListAttachmentsUseCase struct {
	attachmentRepo arepo.AttachmentRepository
	feedbackRepo   frepo.FeedbackRepository
	fileUploader   ports.FileUploader
}

func NewListAttachmentsUseCase(
	attachmentRepo arepo.AttachmentRepository,
	feedbackRepo frepo.FeedbackRepository,
	fileUploader ports.FileUploader,
) *ListAttachmentsUseCase {
	return &ListAttachmentsUseCase{
		attachmentRepo: attachmentRepo,
		feedbackRepo:   feedbackRepo,
		fileUploader:   fileUploader,
	}
}

type ListAttachmentsParams struct {
	FeedbackID   string
	ViewerUserID string
	IsModerator  bool
}

func (uc *ListAttachmentsUseCase) Execute(ctx context.Context, p ListAttachmentsParams) ([]dto.AttachmentResponse, error) {
	f, err := uc.feedbackRepo.FindByID(ctx, p.FeedbackID)
	if err != nil {
		return nil, application.WrapRepoErr(err, fconstant.CodeFeedbackNotFound)
	}
	if f.IsTakedown && !p.IsModerator && f.UserID != p.ViewerUserID {
		return nil, kernel.New(application.ErrCodeNotFound)
	}

	attachments, err := uc.attachmentRepo.ListByFeedbackID(ctx, p.FeedbackID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := make([]dto.AttachmentResponse, 0, len(attachments))
	for _, a := range attachments {
		resp := dto.AttachmentResponse{
			ID:               a.ID,
			Key:              a.Key,
			OriginalFilename: a.OriginalFilename,
			MimeType:         a.MimeType,
			Size:             a.Size,
			CreatedAt:        a.CreatedAt,
		}
		if a.Key != "" {
			url, err := uc.fileUploader.GeneratePresignedDownloadURL(ctx, a.Key, presignExpiry, ports.PrivacyPrivate)
			if err != nil {
				slog.Warn("feedback: presign attachment download gagal", "key", a.Key, "error", err)
			} else {
				resp.DownloadURL = url
			}
		}
		items = append(items, resp)
	}
	return items, nil
}
