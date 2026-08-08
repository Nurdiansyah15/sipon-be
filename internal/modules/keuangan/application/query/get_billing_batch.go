package query

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	bbConst "sipon-be/internal/modules/keuangan/domain/billingbatch/constant"
	bbRepo "sipon-be/internal/modules/keuangan/domain/billingbatch/repository"
)

type GetBillingBatchUseCase struct {
	batchRepo  bbRepo.BillingBatchRepository
	targetRepo bbRepo.BillingBatchTargetRepository
}

func NewGetBillingBatchUseCase(batchRepo bbRepo.BillingBatchRepository, targetRepo bbRepo.BillingBatchTargetRepository) *GetBillingBatchUseCase {
	return &GetBillingBatchUseCase{batchRepo: batchRepo, targetRepo: targetRepo}
}

func (uc *GetBillingBatchUseCase) Execute(ctx context.Context, id string) (*dto.BillingBatchDetailResponse, error) {
	batch, err := uc.batchRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, bbConst.CodeBillingBatchNotFound)
	}

	targets, err := uc.targetRepo.FindByBatchID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, bbConst.CodeBillingBatchQueryFailed)
	}

	resp := &dto.BillingBatchDetailResponse{
		BillingBatchResponse: buildBillingBatchResponse(batch),
		Targets:              make([]dto.BillingBatchTargetResponse, len(targets)),
	}
	for i, t := range targets {
		resp.Targets[i] = buildBillingBatchTargetResponse(t)
	}

	return resp, nil
}
