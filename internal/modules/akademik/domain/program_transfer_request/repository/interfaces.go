package repository

import (
	"context"

	"sipon-be/internal/modules/akademik/domain/program_transfer_request/entity"
)

type ProgramTransferRequestListQuery struct {
	SantriID *string
	Status   *string
	Page     int
	Limit    int
}

type ProgramTransferRequestListResult struct {
	Items []*entity.ProgramTransferRequest
	Total int64
}

type ProgramTransferRequestRepository interface {
	Save(ctx context.Context, req *entity.ProgramTransferRequest) error
	Update(ctx context.Context, req *entity.ProgramTransferRequest) error
	FindByID(ctx context.Context, id string) (*entity.ProgramTransferRequest, error)
	FindPendingBySantriID(ctx context.Context, santriID string) (*entity.ProgramTransferRequest, error)
	List(ctx context.Context, query ProgramTransferRequestListQuery) (*ProgramTransferRequestListResult, error)
}
