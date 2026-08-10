package command

import (
	"context"

	"github.com/google/uuid"

	"sipon-be/internal/modules/feedback/application"
	"sipon-be/internal/modules/feedback/application/dto"
	ports "sipon-be/internal/modules/feedback/application/ports"
	cconstant "sipon-be/internal/modules/feedback/domain/comment/constant"
	crepo "sipon-be/internal/modules/feedback/domain/comment/repository"
	fconstant "sipon-be/internal/modules/feedback/domain/feedback/constant"
	frepo "sipon-be/internal/modules/feedback/domain/feedback/repository"
	lconstant "sipon-be/internal/modules/feedback/domain/like/constant"
	lentity "sipon-be/internal/modules/feedback/domain/like/entity"
	lrepo "sipon-be/internal/modules/feedback/domain/like/repository"
	"sipon-be/internal/shared/kernel"
)

type ToggleLikeUseCase struct {
	likeRepo     lrepo.LikeRepository
	feedbackRepo frepo.FeedbackRepository
	commentRepo  crepo.CommentRepository
	transactor   ports.Transactor
}

func NewToggleLikeUseCase(
	likeRepo lrepo.LikeRepository,
	feedbackRepo frepo.FeedbackRepository,
	commentRepo crepo.CommentRepository,
	transactor ports.Transactor,
) *ToggleLikeUseCase {
	return &ToggleLikeUseCase{
		likeRepo:     likeRepo,
		feedbackRepo: feedbackRepo,
		commentRepo:  commentRepo,
		transactor:   transactor,
	}
}

func (uc *ToggleLikeUseCase) Execute(ctx context.Context, userID string, req dto.ToggleLikeRequest) (*dto.ToggleLikeResponse, error) {
	var targetType lconstant.LikeTargetType
	switch req.TargetType {
	case string(lconstant.TargetFeedback):
		targetType = lconstant.TargetFeedback
	case string(lconstant.TargetComment):
		targetType = lconstant.TargetComment
	default:
		return nil, kernel.New(application.ErrCodeBadRequest)
	}

	if err := uc.validateTargetExists(ctx, targetType, req.TargetID); err != nil {
		return nil, err
	}

	var currentCount int
	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		exists, err := uc.likeRepo.Exists(txCtx, userID, targetType, req.TargetID)
		if err != nil {
			return err
		}

		if exists {
			if err := uc.likeRepo.Delete(txCtx, userID, targetType, req.TargetID); err != nil {
				return err
			}
			switch targetType {
			case lconstant.TargetFeedback:
				if err := uc.feedbackRepo.DecrementLikeCount(txCtx, req.TargetID); err != nil {
					return err
				}
			case lconstant.TargetComment:
				if err := uc.commentRepo.DecrementLikeCount(txCtx, req.TargetID); err != nil {
					return err
				}
			}
		} else {
			like := lentity.NewLike(uuid.NewString(), userID, targetType, req.TargetID)
			if err := uc.likeRepo.Save(txCtx, like); err != nil {
				return err
			}
			switch targetType {
			case lconstant.TargetFeedback:
				if err := uc.feedbackRepo.IncrementLikeCount(txCtx, req.TargetID); err != nil {
					return err
				}
			case lconstant.TargetComment:
				if err := uc.commentRepo.IncrementLikeCount(txCtx, req.TargetID); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	switch targetType {
	case lconstant.TargetFeedback:
		f, err := uc.feedbackRepo.FindByID(ctx, req.TargetID)
		if err != nil {
			return nil, application.WrapRepoErr(err, fconstant.CodeFeedbackNotFound)
		}
		currentCount = f.LikeCount
	case lconstant.TargetComment:
		c, err := uc.commentRepo.FindByID(ctx, req.TargetID)
		if err != nil {
			return nil, application.WrapRepoErr(err, cconstant.CodeCommentNotFound)
		}
		currentCount = c.LikeCount
	}

	liked, err := uc.likeRepo.Exists(ctx, userID, targetType, req.TargetID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.ToggleLikeResponse{Liked: liked, LikeCount: currentCount}, nil
}

func (uc *ToggleLikeUseCase) validateTargetExists(ctx context.Context, targetType lconstant.LikeTargetType, targetID string) error {
	switch targetType {
	case lconstant.TargetFeedback:
		f, err := uc.feedbackRepo.FindByID(ctx, targetID)
		if err != nil {
			return application.WrapRepoErr(err, fconstant.CodeFeedbackNotFound)
		}
		if f.IsTakedown {
			return kernel.New(application.ErrCodeConflict)
		}
	case lconstant.TargetComment:
		c, err := uc.commentRepo.FindByID(ctx, targetID)
		if err != nil {
			return application.WrapRepoErr(err, cconstant.CodeCommentNotFound)
		}
		if c.IsTakedown {
			return kernel.New(application.ErrCodeConflict)
		}
	}
	return nil
}
