package command

import (
	"context"

	"sipon-be/internal/modules/feedback/application"
	"sipon-be/internal/modules/feedback/application/dto"
	cconstant "sipon-be/internal/modules/feedback/domain/comment/constant"
	crepo "sipon-be/internal/modules/feedback/domain/comment/repository"
	frepo "sipon-be/internal/modules/feedback/domain/feedback/repository"
	"sipon-be/internal/shared/kernel"
)

type DeleteCommentUseCase struct {
	commentRepo  crepo.CommentRepository
	feedbackRepo frepo.FeedbackRepository
}

func NewDeleteCommentUseCase(commentRepo crepo.CommentRepository, feedbackRepo frepo.FeedbackRepository) *DeleteCommentUseCase {
	return &DeleteCommentUseCase{commentRepo: commentRepo, feedbackRepo: feedbackRepo}
}

func (uc *DeleteCommentUseCase) Execute(ctx context.Context, userID, commentID string) (*dto.MessageResponse, error) {
	c, err := uc.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		return nil, application.WrapRepoErr(err, cconstant.CodeCommentNotFound)
	}

	if c.UserID != userID {
		return nil, kernel.New(application.ErrCodeForbidden)
	}

	c.SoftDelete()
	if err := uc.commentRepo.Update(ctx, c); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if f, err := uc.feedbackRepo.FindByID(ctx, c.FeedbackID); err == nil && !f.IsTakedown {
		f.DecrementComment()
		_ = uc.feedbackRepo.Update(ctx, f)
	}

	return &dto.MessageResponse{Message: "komentar berhasil dihapus"}, nil
}
