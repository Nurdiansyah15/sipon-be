package command

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

type DeleteFeedbackUseCase struct {
	feedbackRepo   frepo.FeedbackRepository
	attachmentRepo arepo.AttachmentRepository
	fileUploader   ports.FileUploader
	transactor     ports.Transactor
}

func NewDeleteFeedbackUseCase(
	feedbackRepo frepo.FeedbackRepository,
	attachmentRepo arepo.AttachmentRepository,
	fileUploader ports.FileUploader,
	transactor ports.Transactor,
) *DeleteFeedbackUseCase {
	return &DeleteFeedbackUseCase{
		feedbackRepo:   feedbackRepo,
		attachmentRepo: attachmentRepo,
		fileUploader:   fileUploader,
		transactor:     transactor,
	}
}

func (uc *DeleteFeedbackUseCase) Execute(ctx context.Context, userID, feedbackID string) (*dto.MessageResponse, error) {
	f, err := uc.feedbackRepo.FindByID(ctx, feedbackID)
	if err != nil {
		return nil, application.WrapRepoErr(err, fconstant.CodeFeedbackNotFound)
	}

	if f.UserID != userID {
		return nil, kernel.New(application.ErrCodeForbidden)
	}

	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		f.SoftDelete()
		if err := uc.feedbackRepo.Update(txCtx, f); err != nil {
			return err
		}
		attachments, err := uc.attachmentRepo.ListByFeedbackID(txCtx, feedbackID)
		if err != nil {
			return err
		}
		for _, a := range attachments {
			if a.DeletedAt != nil {
				continue
			}
			if err := uc.attachmentRepo.SoftDelete(txCtx, a.ID); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	go func() {
		attachments, err := uc.attachmentRepo.ListByFeedbackID(context.Background(), feedbackID)
		if err != nil {
			slog.Warn("feedback: gagal list attachment untuk cleanup", "feedback_id", feedbackID, "error", err)
			return
		}
		for _, a := range attachments {
			if err := uc.fileUploader.DeleteObject(context.Background(), a.Key, ports.PrivacyPrivate); err != nil {
				slog.Warn("feedback: gagal hapus objek attachment", "key", a.Key, "error", err)
			}
		}
	}()

	return &dto.MessageResponse{Message: "feedback berhasil dihapus"}, nil
}
