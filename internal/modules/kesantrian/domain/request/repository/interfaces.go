package repository

import (
	"context"

	"sipon-be/internal/modules/kesantrian/domain/request/constant"
	"sipon-be/internal/modules/kesantrian/domain/request/entity"
)

type SantriRequestListQuery struct {
	Status   *constant.SantriRequestStatus
	Page     int
	Limit    int
	SortBy   string
	SortType string
}

type SantriRequestListResult struct {
	Items []*entity.SantriRequest
	Total int64
}

type SantriRequestRepository interface {
	Save(ctx context.Context, request *entity.SantriRequest) error
	Update(ctx context.Context, request *entity.SantriRequest) error
	FindByID(ctx context.Context, id string) (*entity.SantriRequest, error)
	FindPendingByUserID(ctx context.Context, userID string) (*entity.SantriRequest, error)
	List(ctx context.Context, query SantriRequestListQuery) (*SantriRequestListResult, error)
}
