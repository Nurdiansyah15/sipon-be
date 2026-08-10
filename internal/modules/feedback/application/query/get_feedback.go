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
	lconstant "sipon-be/internal/modules/feedback/domain/like/constant"
	lrepo "sipon-be/internal/modules/feedback/domain/like/repository"
)

type GetFeedbackUseCase struct {
	feedbackRepo   frepo.FeedbackRepository
	attachmentRepo arepo.AttachmentRepository
	likeRepo       lrepo.LikeRepository
	identityReader ports.IdentityReader
	fileUploader   ports.FileUploader
}

func NewGetFeedbackUseCase(
	feedbackRepo frepo.FeedbackRepository,
	attachmentRepo arepo.AttachmentRepository,
	likeRepo lrepo.LikeRepository,
	identityReader ports.IdentityReader,
	fileUploader ports.FileUploader,
) *GetFeedbackUseCase {
	return &GetFeedbackUseCase{
		feedbackRepo:   feedbackRepo,
		attachmentRepo: attachmentRepo,
		likeRepo:       likeRepo,
		identityReader: identityReader,
		fileUploader:   fileUploader,
	}
}

type GetFeedbackParams struct {
	FeedbackID   string
	ViewerUserID string
	IsModerator  bool
}

func (uc *GetFeedbackUseCase) Execute(ctx context.Context, p GetFeedbackParams) (*dto.FeedbackDetailResponse, error) {
	f, err := uc.feedbackRepo.FindByID(ctx, p.FeedbackID)
	if err != nil {
		return nil, application.WrapRepoErr(err, fconstant.CodeFeedbackNotFound)
	}

	// Takedown feedback is only visible to its owner or a moderator.
	if f.IsTakedown && !p.IsModerator && f.UserID != p.ViewerUserID {
		return nil, kernelErrNotFound()
	}

	users := enrichUsers(ctx, uc.identityReader, []string{f.UserID})
	likes := likedTargets(ctx, uc.likeRepo, p.ViewerUserID, lconstant.TargetFeedback, []string{f.ID})

	attachments, err := uc.attachmentRepo.ListByFeedbackID(ctx, p.FeedbackID)
	if err != nil {
		return nil, kernelErrInternal()
	}
	attachmentResp := make([]dto.AttachmentResponse, 0, len(attachments))
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
		attachmentResp = append(attachmentResp, resp)
	}

	return &dto.FeedbackDetailResponse{
		ListFeedbackItem: dto.ListFeedbackItem{
			ID:              f.ID,
			User:            users[f.UserID],
			Title:           f.Title,
			Body:            f.Body,
			Category:        string(f.Category),
			IsTakedown:      f.IsTakedown,
			TakedownReason:  f.TakedownReason,
			LikeCount:       f.LikeCount,
			CommentCount:    f.CommentCount,
			IsLiked:         likes[f.ID],
			AttachmentCount: len(attachmentResp),
			CreatedAt:       f.CreatedAt,
			UpdatedAt:       f.UpdatedAt,
		},
		Attachments: attachmentResp,
		IsOwner:     f.UserID == p.ViewerUserID,
	}, nil
}
