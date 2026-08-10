package query

import (
	"context"
	"log/slog"
	"time"

	"sipon-be/internal/modules/feedback/application"
	"sipon-be/internal/modules/feedback/application/dto"
	ports "sipon-be/internal/modules/feedback/application/ports"
	lconstant "sipon-be/internal/modules/feedback/domain/like/constant"
	lrepo "sipon-be/internal/modules/feedback/domain/like/repository"
	"sipon-be/internal/shared/kernel"
)

const presignExpiry = 15 * time.Minute

func kernelErrNotFound() error {
	return kernel.New(application.ErrCodeNotFound)
}

func kernelErrInternal() error {
	return kernel.New(application.ErrCodeInternal)
}

// enrichUsers resolves user summaries for a set of user IDs (N+1 by design —
// consistent with kesantrian; enrichment failure is logged, not fatal).
func enrichUsers(ctx context.Context, reader ports.IdentityReader, userIDs []string) map[string]*dto.UserSummaryDTO {
	result := make(map[string]*dto.UserSummaryDTO)
	for _, userID := range userIDs {
		if userID == "" {
			continue
		}
		if _, done := result[userID]; done {
			continue
		}
		summary, err := reader.GetUserSummary(ctx, userID)
		if err != nil {
			slog.Warn("feedback: user summary enrichment failed", "user_id", userID, "error", err)
			continue
		}
		result[userID] = &dto.UserSummaryDTO{
			UserID:   summary.UserID,
			Username: summary.Username,
			Email:    summary.Email,
			Fullname: summary.Fullname,
		}
	}
	return result
}

// likedTargets returns which of the given target IDs the user has liked.
func likedTargets(ctx context.Context, likeRepo lrepo.LikeRepository, userID string, targetType lconstant.LikeTargetType, targetIDs []string) map[string]bool {
	if userID == "" || len(targetIDs) == 0 {
		return map[string]bool{}
	}
	res, err := likeRepo.ListLikedTargetIDs(ctx, userID, targetType, targetIDs)
	if err != nil {
		slog.Warn("feedback: liked targets query failed", "error", err)
		return map[string]bool{}
	}
	return res
}
