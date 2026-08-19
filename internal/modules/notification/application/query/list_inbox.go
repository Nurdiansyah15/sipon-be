package query

import (
	"context"

	"sipon-be/internal/modules/notification/application/dto"
	"sipon-be/internal/modules/notification/domain/delivery/repository"
)

type ListInboxUseCase struct {
	deliveryRepo repository.DeliveryAttemptRepository
}

func NewListInboxUseCase(repo repository.DeliveryAttemptRepository) *ListInboxUseCase {
	return &ListInboxUseCase{deliveryRepo: repo}
}

func (uc *ListInboxUseCase) Execute(ctx context.Context, userID string, req dto.ListNotificationsQuery) ([]dto.NotificationItem, repository.Meta, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 || limit > 50 {
		limit = 20
	}

	items, meta, err := uc.deliveryRepo.ListInApp(ctx, repository.ListInAppQuery{
		UserID:     userID,
		UnreadOnly: req.UnreadOnly,
		Page:       page,
		Limit:      limit,
	})
	if err != nil {
		return nil, repository.Meta{}, err
	}

	result := make([]dto.NotificationItem, 0, len(items))
	for _, item := range items {
		n := dto.NotificationItem{
			ID:          item.DeliveryAttemptID,
			Type:        item.Type,
			Title:       item.Title,
			Body:        item.Body,
			ImageURL:    resolveMediaURL(item.ImageURL),
			Module:      item.Module,
			EventType:   item.EventType,
			EntityID:    item.EntityID,
			ClickAction: item.ClickAction,
			Bypass:      item.Bypass,
			Extra:       item.Extra,
			IsRead:      item.ReadAt != nil,
			CreatedAt:   dto.FormatRFC3339(item.AttemptedAt),
		}
		if item.ReadAt != nil {
			s := dto.FormatRFC3339(*item.ReadAt)
			n.ReadAt = &s
		}
		result = append(result, n)
	}

	return result, meta, nil
}

func resolveMediaURL(url string) string {
	return url
}
