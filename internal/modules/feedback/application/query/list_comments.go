package query

import (
	"context"
	"math"

	"sipon-be/internal/modules/feedback/application"
	"sipon-be/internal/modules/feedback/application/dto"
	ports "sipon-be/internal/modules/feedback/application/ports"
	crepo "sipon-be/internal/modules/feedback/domain/comment/repository"
	fconstant "sipon-be/internal/modules/feedback/domain/feedback/constant"
	frepo "sipon-be/internal/modules/feedback/domain/feedback/repository"
	lconstant "sipon-be/internal/modules/feedback/domain/like/constant"
	lrepo "sipon-be/internal/modules/feedback/domain/like/repository"
	"sipon-be/internal/shared/kernel"
)

type ListCommentsUseCase struct {
	commentRepo    crepo.CommentRepository
	feedbackRepo   frepo.FeedbackRepository
	likeRepo       lrepo.LikeRepository
	identityReader ports.IdentityReader
}

func NewListCommentsUseCase(
	commentRepo crepo.CommentRepository,
	feedbackRepo frepo.FeedbackRepository,
	likeRepo lrepo.LikeRepository,
	identityReader ports.IdentityReader,
) *ListCommentsUseCase {
	return &ListCommentsUseCase{
		commentRepo:    commentRepo,
		feedbackRepo:   feedbackRepo,
		likeRepo:       likeRepo,
		identityReader: identityReader,
	}
}

type ListCommentsParams struct {
	FeedbackID      string
	ViewerUserID    string
	IncludeTakedown bool
	Page            int
	Limit           int
}

func (uc *ListCommentsUseCase) Execute(ctx context.Context, p ListCommentsParams) ([]dto.CommentItemResponse, *dto.PaginationMeta, error) {
	f, err := uc.feedbackRepo.FindByID(ctx, p.FeedbackID)
	if err != nil {
		return nil, nil, application.WrapRepoErr(err, fconstant.CodeFeedbackNotFound)
	}

	// Takedown feedback's comments are only visible to owner or moderator
	// (moderator is implied by IncludeTakedown here since only admins list
	// takedown comments; owner check is done per-feedback).
	if f.IsTakedown && !p.IncludeTakedown && f.UserID != p.ViewerUserID {
		return nil, nil, kernel.New(application.ErrCodeNotFound)
	}

	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 || p.Limit > 100 {
		p.Limit = 10
	}

	result, err := uc.commentRepo.List(ctx, crepo.CommentListQuery{
		FeedbackID:      p.FeedbackID,
		IncludeTakedown: p.IncludeTakedown,
		Page:            p.Page,
		Limit:           p.Limit,
	})
	if err != nil {
		return nil, nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := make([]dto.CommentItemResponse, len(result.Items))
	userIDs := make([]string, 0, len(result.Items)*2)
	commentIDs := make([]string, 0, len(result.Items))
	for _, c := range result.Items {
		userIDs = append(userIDs, c.UserID)
		commentIDs = append(commentIDs, c.ID)
		if c.ReplyToID != nil {
			userIDs = append(userIDs, *c.ReplyToID)
		}
	}

	users := enrichUsers(ctx, uc.identityReader, userIDs)
	likes := likedTargets(ctx, uc.likeRepo, p.ViewerUserID, lconstant.TargetComment, commentIDs)

	// Resolve the original author of each reply_to comment to display
	// "replying to <name>". N+1 by design (same convention as user summary).
	replyToAuthor := make(map[string]*dto.UserSummaryDTO)
	for _, c := range result.Items {
		if c.ReplyToID == nil {
			continue
		}
		rc, err := uc.commentRepo.FindByID(ctx, *c.ReplyToID)
		if err != nil || rc.DeletedAt != nil {
			continue
		}
		replyToAuthor[*c.ReplyToID] = users[rc.UserID]
	}

	for i, c := range result.Items {
		var replyToUser *dto.UserSummaryDTO
		if c.ReplyToID != nil {
			replyToUser = replyToAuthor[*c.ReplyToID]
		}
		items[i] = dto.CommentItemResponse{
			ID:             c.ID,
			FeedbackID:     c.FeedbackID,
			User:           users[c.UserID],
			Body:           c.Body,
			ReplyToID:      c.ReplyToID,
			ReplyToUser:    replyToUser,
			IsTakedown:     c.IsTakedown,
			TakedownReason: c.TakedownReason,
			LikeCount:      c.LikeCount,
			IsLiked:        likes[c.ID],
			IsOwner:        c.UserID == p.ViewerUserID,
			CreatedAt:      c.CreatedAt,
			UpdatedAt:      c.UpdatedAt,
		}
	}

	totalPages := int(math.Ceil(float64(result.Total) / float64(p.Limit)))
	meta := &dto.PaginationMeta{
		CurrentPage: p.Page,
		PerPage:     p.Limit,
		Total:       int(result.Total),
		TotalPages:  totalPages,
	}

	return items, meta, nil
}
