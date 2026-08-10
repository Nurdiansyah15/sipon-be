package command

import (
	"context"

	"github.com/google/uuid"

	"sipon-be/internal/modules/feedback/application"
	"sipon-be/internal/modules/feedback/application/dto"
	ports "sipon-be/internal/modules/feedback/application/ports"
	cconstant "sipon-be/internal/modules/feedback/domain/comment/constant"
	centity "sipon-be/internal/modules/feedback/domain/comment/entity"
	crepo "sipon-be/internal/modules/feedback/domain/comment/repository"
	fconstant "sipon-be/internal/modules/feedback/domain/feedback/constant"
	frepo "sipon-be/internal/modules/feedback/domain/feedback/repository"
	"sipon-be/internal/shared/kernel"
)

type CreateCommentUseCase struct {
	commentRepo  crepo.CommentRepository
	feedbackRepo frepo.FeedbackRepository
	transactor   ports.Transactor
}

func NewCreateCommentUseCase(commentRepo crepo.CommentRepository, feedbackRepo frepo.FeedbackRepository, transactor ports.Transactor) *CreateCommentUseCase {
	return &CreateCommentUseCase{
		commentRepo:  commentRepo,
		feedbackRepo: feedbackRepo,
		transactor:   transactor,
	}
}

func (uc *CreateCommentUseCase) Execute(ctx context.Context, userID, feedbackID string, req dto.CreateCommentRequest) (*dto.MessageResponse, error) {
	f, err := uc.feedbackRepo.FindByID(ctx, feedbackID)
	if err != nil {
		return nil, application.WrapRepoErr(err, fconstant.CodeFeedbackNotFound)
	}
	if f.IsTakedown {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	if req.ReplyToID != nil && *req.ReplyToID != "" {
		replyTo, err := uc.commentRepo.FindByID(ctx, *req.ReplyToID)
		if err != nil {
			return nil, application.WrapRepoErr(err, cconstant.CodeCommentNotFound)
		}
		if replyTo.FeedbackID != feedbackID || replyTo.DeletedAt != nil || replyTo.IsTakedown {
			return nil, kernel.New(application.ErrCodeBadRequest)
		}
	}

	c, err := centity.NewComment(uuid.NewString(), feedbackID, userID, req.Body, req.ReplyToID)
	if err != nil {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := uc.commentRepo.Save(txCtx, c); err != nil {
			return err
		}
		f.IncrementComment()
		return uc.feedbackRepo.Update(txCtx, f)
	}); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "komentar berhasil ditambahkan"}, nil
}
