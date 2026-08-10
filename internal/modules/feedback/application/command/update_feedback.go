package command

import (
	"context"

	"sipon-be/internal/modules/feedback/application"
	"sipon-be/internal/modules/feedback/application/dto"
	fconstant "sipon-be/internal/modules/feedback/domain/feedback/constant"
	frepo "sipon-be/internal/modules/feedback/domain/feedback/repository"
	"sipon-be/internal/shared/kernel"
)

type UpdateFeedbackUseCase struct {
	feedbackRepo frepo.FeedbackRepository
}

func NewUpdateFeedbackUseCase(feedbackRepo frepo.FeedbackRepository) *UpdateFeedbackUseCase {
	return &UpdateFeedbackUseCase{feedbackRepo: feedbackRepo}
}

func (uc *UpdateFeedbackUseCase) Execute(ctx context.Context, userID, feedbackID string, req dto.UpdateFeedbackRequest) (*dto.MessageResponse, error) {
	f, err := uc.feedbackRepo.FindByID(ctx, feedbackID)
	if err != nil {
		return nil, application.WrapRepoErr(err, fconstant.CodeFeedbackNotFound)
	}

	if f.UserID != userID {
		return nil, kernel.New(application.ErrCodeForbidden)
	}

	if err := f.Update(req.Title, req.Body, fconstant.FeedbackCategory(req.Category)); err != nil {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	if err := uc.feedbackRepo.Update(ctx, f); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "feedback berhasil diupdate"}, nil
}
