package command

import (
	"context"

	"sipon-be/internal/modules/feedback/application"
	"sipon-be/internal/modules/feedback/application/dto"
	cconstant "sipon-be/internal/modules/feedback/domain/comment/constant"
	crepo "sipon-be/internal/modules/feedback/domain/comment/repository"
	"sipon-be/internal/shared/kernel"
)

type UpdateCommentUseCase struct {
	commentRepo crepo.CommentRepository
}

func NewUpdateCommentUseCase(commentRepo crepo.CommentRepository) *UpdateCommentUseCase {
	return &UpdateCommentUseCase{commentRepo: commentRepo}
}

func (uc *UpdateCommentUseCase) Execute(ctx context.Context, userID, commentID string, req dto.UpdateCommentRequest) (*dto.MessageResponse, error) {
	c, err := uc.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		return nil, application.WrapRepoErr(err, cconstant.CodeCommentNotFound)
	}

	if c.UserID != userID {
		return nil, kernel.New(application.ErrCodeForbidden)
	}

	if err := c.Update(req.Body); err != nil {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	if err := uc.commentRepo.Update(ctx, c); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "komentar berhasil diupdate"}, nil
}
