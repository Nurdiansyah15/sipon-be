package query

import (
	"context"
	"math"
	"strings"

	"sipon-be/internal/modules/feedback/application"
	"sipon-be/internal/modules/feedback/application/dto"
	ports "sipon-be/internal/modules/feedback/application/ports"
	arepo "sipon-be/internal/modules/feedback/domain/attachment/repository"
	frepo "sipon-be/internal/modules/feedback/domain/feedback/repository"
	lconstant "sipon-be/internal/modules/feedback/domain/like/constant"
	lrepo "sipon-be/internal/modules/feedback/domain/like/repository"
	"sipon-be/internal/shared/kernel"
)

type ListFeedbacksUseCase struct {
	feedbackRepo   frepo.FeedbackRepository
	attachmentRepo arepo.AttachmentRepository
	likeRepo       lrepo.LikeRepository
	identityReader ports.IdentityReader
}

func NewListFeedbacksUseCase(
	feedbackRepo frepo.FeedbackRepository,
	attachmentRepo arepo.AttachmentRepository,
	likeRepo lrepo.LikeRepository,
	identityReader ports.IdentityReader,
) *ListFeedbacksUseCase {
	return &ListFeedbacksUseCase{
		feedbackRepo:   feedbackRepo,
		attachmentRepo: attachmentRepo,
		likeRepo:       likeRepo,
		identityReader: identityReader,
	}
}

type ListFeedbacksParams struct {
	Query           dto.ListFeedbackQuery
	ViewerUserID    string
	IncludeTakedown bool
	OnlyMine        bool
}

func (uc *ListFeedbacksUseCase) Execute(ctx context.Context, p ListFeedbacksParams) ([]dto.ListFeedbackItem, *dto.PaginationMeta, error) {
	if p.Query.Page < 1 {
		p.Query.Page = 1
	}
	if p.Query.Limit < 1 || p.Query.Limit > 100 {
		p.Query.Limit = 10
	}

	var categoryPtr *string
	if p.Query.Category != "" {
		c := p.Query.Category
		categoryPtr = &c
	}

	ownerID := ""
	if p.OnlyMine {
		ownerID = p.ViewerUserID
	}

	result, err := uc.feedbackRepo.List(ctx, frepo.FeedbackListQuery{
		Category:        categoryPtr,
		Search:          strings.TrimSpace(p.Query.Search),
		UserID:          ownerID,
		IncludeTakedown: p.IncludeTakedown,
		Page:            p.Query.Page,
		Limit:           p.Query.Limit,
	})
	if err != nil {
		return nil, nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := make([]dto.ListFeedbackItem, len(result.Items))
	userIDs := make([]string, 0, len(result.Items))
	feedbackIDs := make([]string, 0, len(result.Items))
	for _, f := range result.Items {
		userIDs = append(userIDs, f.UserID)
		feedbackIDs = append(feedbackIDs, f.ID)
	}

	users := enrichUsers(ctx, uc.identityReader, userIDs)
	likes := likedTargets(ctx, uc.likeRepo, p.ViewerUserID, lconstant.TargetFeedback, feedbackIDs)
	attachmentCounts, _ := uc.attachmentRepo.CountByFeedbackIDs(ctx, feedbackIDs)

	for i, f := range result.Items {
		item := dto.ListFeedbackItem{
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
			AttachmentCount: int(attachmentCounts[f.ID]),
			CreatedAt:       f.CreatedAt,
			UpdatedAt:       f.UpdatedAt,
		}
		items[i] = item
	}

	totalPages := int(math.Ceil(float64(result.Total) / float64(p.Query.Limit)))
	meta := &dto.PaginationMeta{
		CurrentPage: p.Query.Page,
		PerPage:     p.Query.Limit,
		Total:       int(result.Total),
		TotalPages:  totalPages,
	}

	return items, meta, nil
}
