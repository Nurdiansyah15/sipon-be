package command

import (
	"context"

	"sipon-be/internal/modules/feedback/application"
	"sipon-be/internal/modules/feedback/application/dto"
	fconstant "sipon-be/internal/modules/feedback/domain/feedback/constant"
	frepo "sipon-be/internal/modules/feedback/domain/feedback/repository"
	"sipon-be/internal/shared/kernel"
)

type ModerateFeedbackUseCase struct {
	feedbackRepo frepo.FeedbackRepository
}

func NewModerateFeedbackUseCase(feedbackRepo frepo.FeedbackRepository) *ModerateFeedbackUseCase {
	return &ModerateFeedbackUseCase{feedbackRepo: feedbackRepo}
}

type ModerateFeedbackRequest struct {
	Reason *string `json:"reason,omitempty"`
}

func (uc *ModerateFeedbackUseCase) Takedown(ctx context.Context, adminID, feedbackID string, req ModerateFeedbackRequest) (*dto.MessageResponse, error) {
	f, err := uc.feedbackRepo.FindByID(ctx, feedbackID)
	if err != nil {
		return nil, application.WrapRepoErr(err, fconstant.CodeFeedbackNotFound)
	}

	if err := f.Takedown(adminID, req.Reason); err != nil {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	if err := uc.feedbackRepo.Update(ctx, f); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "feedback berhasil ditakedown"}, nil
}

func (uc *ModerateFeedbackUseCase) Restore(ctx context.Context, feedbackID string) (*dto.MessageResponse, error) {
	f, err := uc.feedbackRepo.FindByID(ctx, feedbackID)
	if err != nil {
		return nil, application.WrapRepoErr(err, fconstant.CodeFeedbackNotFound)
	}

	if err := f.Restore(); err != nil {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	if err := uc.feedbackRepo.Update(ctx, f); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "feedback berhasil direstore"}, nil
}
