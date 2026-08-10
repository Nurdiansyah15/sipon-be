package command

import (
	"context"

	"github.com/google/uuid"

	"sipon-be/internal/modules/feedback/application"
	"sipon-be/internal/modules/feedback/application/dto"
	ports "sipon-be/internal/modules/feedback/application/ports"
	fconstant "sipon-be/internal/modules/feedback/domain/feedback/constant"
	fentity "sipon-be/internal/modules/feedback/domain/feedback/entity"
	frepo "sipon-be/internal/modules/feedback/domain/feedback/repository"
	"sipon-be/internal/shared/kernel"
)

type CreateFeedbackUseCase struct {
	feedbackRepo frepo.FeedbackRepository
	transactor   ports.Transactor
}

func NewCreateFeedbackUseCase(feedbackRepo frepo.FeedbackRepository, transactor ports.Transactor) *CreateFeedbackUseCase {
	return &CreateFeedbackUseCase{feedbackRepo: feedbackRepo, transactor: transactor}
}

func (uc *CreateFeedbackUseCase) Execute(ctx context.Context, userID string, req dto.CreateFeedbackRequest) (*dto.FeedbackDetailResponse, error) {
	f, err := fentity.NewFeedback(uuid.NewString(), userID, req.Title, req.Body, fconstant.FeedbackCategory(req.Category))
	if err != nil {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		return uc.feedbackRepo.Save(txCtx, f)
	}); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.FeedbackDetailResponse{
		ListFeedbackItem: dto.ListFeedbackItem{
			ID:         f.ID,
			Title:      f.Title,
			Body:       f.Body,
			Category:   string(f.Category),
			IsTakedown: f.IsTakedown,
			LikeCount:  f.LikeCount,
			IsLiked:    false,
			CreatedAt:  f.CreatedAt,
			UpdatedAt:  f.UpdatedAt,
		},
		Attachments: []dto.AttachmentResponse{},
		IsOwner:     true,
	}, nil
}
