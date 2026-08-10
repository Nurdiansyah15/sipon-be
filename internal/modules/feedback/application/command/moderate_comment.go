package command

import (
	"context"

	"sipon-be/internal/modules/feedback/application"
	"sipon-be/internal/modules/feedback/application/dto"
	cconstant "sipon-be/internal/modules/feedback/domain/comment/constant"
	crepo "sipon-be/internal/modules/feedback/domain/comment/repository"
	"sipon-be/internal/shared/kernel"
)

type ModerateCommentUseCase struct {
	commentRepo crepo.CommentRepository
}

func NewModerateCommentUseCase(commentRepo crepo.CommentRepository) *ModerateCommentUseCase {
	return &ModerateCommentUseCase{commentRepo: commentRepo}
}

type ModerateCommentRequest struct {
	Reason *string `json:"reason,omitempty"`
}

func (uc *ModerateCommentUseCase) Takedown(ctx context.Context, adminID, commentID string, req ModerateCommentRequest) (*dto.MessageResponse, error) {
	c, err := uc.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		return nil, application.WrapRepoErr(err, cconstant.CodeCommentNotFound)
	}

	if err := c.Takedown(adminID, req.Reason); err != nil {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	if err := uc.commentRepo.Update(ctx, c); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "komentar berhasil ditakedown"}, nil
}

func (uc *ModerateCommentUseCase) Restore(ctx context.Context, commentID string) (*dto.MessageResponse, error) {
	c, err := uc.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		return nil, application.WrapRepoErr(err, cconstant.CodeCommentNotFound)
	}

	if err := c.Restore(); err != nil {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	if err := uc.commentRepo.Update(ctx, c); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "komentar berhasil direstore"}, nil
}
