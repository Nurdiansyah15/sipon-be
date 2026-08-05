package command

import (
	"context"

	"github.com/google/uuid"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	feeConst "sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	feeEntity "sipon-be/internal/modules/keuangan/domain/feecomponent/entity"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	"sipon-be/internal/shared/kernel"
)

type CreateFeeComponentUseCase struct {
	feeComponentRepo feeRepo.FeeComponentRepository
}

func NewCreateFeeComponentUseCase(feeComponentRepo feeRepo.FeeComponentRepository) *CreateFeeComponentUseCase {
	return &CreateFeeComponentUseCase{feeComponentRepo: feeComponentRepo}
}

func (uc *CreateFeeComponentUseCase) Execute(ctx context.Context, req dto.CreateFeeComponentRequest, createdBy string) (*dto.FeeComponentResponse, error) {
	exists, err := uc.feeComponentRepo.ExistsByCode(ctx, req.Code, "")
	if err != nil {
		return nil, application.WrapRepoErr(err, feeConst.CodeFeeComponentNotFound)
	}
	if exists {
		return nil, kernel.New(feeConst.CodeFeeComponentDuplicate)
	}

	feeType := feeConst.FeeComponentType(req.Type)
	if !feeConst.IsValidFeeType(feeType) {
		return nil, kernel.New(feeConst.CodeFeeComponentInvalidType)
	}

	var periodType *feeConst.PeriodType
	if req.PeriodType != nil {
		pt := feeConst.PeriodType(*req.PeriodType)
		periodType = &pt
	}

	fc, err := feeEntity.NewFeeComponent(uuid.New().String(), req.Code, req.Name, feeType, req.Amount, createdBy)
	if err != nil {
		return nil, application.WrapRepoErr(err, feeConst.CodeFeeComponentNotFound)
	}
	fc.IsPeriodic = req.IsPeriodic
	fc.PeriodType = periodType
	fc.Description = req.Description

	if err := uc.feeComponentRepo.Save(ctx, fc); err != nil {
		return nil, application.WrapRepoErr(err, feeConst.CodeFeeComponentNotFound)
	}

	return toFeeComponentResponse(fc), nil
}

func toFeeComponentResponse(fc *feeEntity.FeeComponent) *dto.FeeComponentResponse {
	resp := &dto.FeeComponentResponse{
		ID:         fc.ID,
		Code:       fc.Code,
		Name:       fc.Name,
		Type:       string(fc.Type),
		Amount:     fc.Amount,
		IsPeriodic: fc.IsPeriodic,
		IsActive:   fc.IsActive,
		CreatedAt:  fc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  fc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if fc.PeriodType != nil {
		s := string(*fc.PeriodType)
		resp.PeriodType = &s
	}
	if fc.Description != nil {
		resp.Description = fc.Description
	}
	return resp
}
