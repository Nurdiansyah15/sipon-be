package command

import (
	"context"

	"github.com/google/uuid"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	bsConst "sipon-be/internal/modules/keuangan/domain/billingscheme/constant"
	bsEntity "sipon-be/internal/modules/keuangan/domain/billingscheme/entity"
	bsRepo "sipon-be/internal/modules/keuangan/domain/billingscheme/repository"
)

type CreateBillingSchemeUseCase struct {
	billingSchemeRepo bsRepo.BillingSchemeRepository
}

func NewCreateBillingSchemeUseCase(billingSchemeRepo bsRepo.BillingSchemeRepository) *CreateBillingSchemeUseCase {
	return &CreateBillingSchemeUseCase{billingSchemeRepo: billingSchemeRepo}
}

func (uc *CreateBillingSchemeUseCase) Execute(ctx context.Context, req dto.CreateBillingSchemeRequest, createdBy string) (*dto.BillingSchemeResponse, error) {
	scheme, err := bsEntity.NewBillingScheme(uuid.New().String(), req.Name, createdBy)
	if err != nil {
		return nil, application.WrapRepoErr(err, bsConst.CodeBillingSchemeNotFound)
	}
	scheme.Description = req.Description

	if err := uc.billingSchemeRepo.Save(ctx, scheme); err != nil {
		return nil, application.WrapRepoErr(err, bsConst.CodeBillingSchemeNotFound)
	}

	return toBillingSchemeResponse(scheme), nil
}

func toBillingSchemeResponse(scheme *bsEntity.BillingScheme) *dto.BillingSchemeResponse {
	resp := &dto.BillingSchemeResponse{
		ID:          scheme.ID,
		Name:        scheme.Name,
		Description: scheme.Description,
		IsActive:    scheme.IsActive,
		CreatedAt:   scheme.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   scheme.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if scheme.Items != nil {
		resp.Items = make([]dto.BillingSchemeItemResponse, len(scheme.Items))
		for i, item := range scheme.Items {
			resp.Items[i] = dto.BillingSchemeItemResponse{
				ID:             item.ID,
				FeeComponentID: item.FeeComponentID,
				AmountOverride: item.AmountOverride,
				IsRequired:     item.IsRequired,
				SortOrder:      item.SortOrder,
			}
		}
	}
	return resp
}
