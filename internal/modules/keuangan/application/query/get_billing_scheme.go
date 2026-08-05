package query

import (
	"context"

	bsConst "sipon-be/internal/modules/keuangan/domain/billingscheme/constant"
	bsRepo "sipon-be/internal/modules/keuangan/domain/billingscheme/repository"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
)

type GetBillingSchemeUseCase struct {
	billingSchemeRepo bsRepo.BillingSchemeRepository
	feeComponentRepo  feeRepo.FeeComponentRepository
}

func NewGetBillingSchemeUseCase(
	billingSchemeRepo bsRepo.BillingSchemeRepository,
	feeComponentRepo feeRepo.FeeComponentRepository,
) *GetBillingSchemeUseCase {
	return &GetBillingSchemeUseCase{
		billingSchemeRepo: billingSchemeRepo,
		feeComponentRepo:  feeComponentRepo,
	}
}

func (uc *GetBillingSchemeUseCase) Execute(ctx context.Context, id string) (*dto.BillingSchemeResponse, error) {
	bs, err := uc.billingSchemeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, bsConst.CodeBillingSchemeNotFound)
	}

	resp := &dto.BillingSchemeResponse{
		ID:          bs.ID,
		Name:        bs.Name,
		Description: bs.Description,
		IsActive:    bs.IsActive,
		CreatedAt:   bs.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   bs.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if bs.Items != nil {
		resp.Items = make([]dto.BillingSchemeItemResponse, len(bs.Items))
		for i, item := range bs.Items {
			itemResp := dto.BillingSchemeItemResponse{
				ID:             item.ID,
				FeeComponentID: item.FeeComponentID,
				AmountOverride: item.AmountOverride,
				IsRequired:     item.IsRequired,
				SortOrder:      item.SortOrder,
			}

			// Fetch fee component details
			fc, err := uc.feeComponentRepo.FindByID(ctx, item.FeeComponentID)
			if err == nil {
				itemResp.FeeComponent = &dto.FeeComponentBriefResponse{
					ID:     fc.ID,
					Code:   fc.Code,
					Name:   fc.Name,
					Type:   string(fc.Type),
					Amount: fc.Amount,
				}
			}

			resp.Items[i] = itemResp
		}
	}

	return resp, nil
}
