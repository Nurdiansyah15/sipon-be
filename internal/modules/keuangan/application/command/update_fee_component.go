package command

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	feeConst "sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
)

type UpdateFeeComponentUseCase struct {
	feeComponentRepo feeRepo.FeeComponentRepository
}

func NewUpdateFeeComponentUseCase(feeComponentRepo feeRepo.FeeComponentRepository) *UpdateFeeComponentUseCase {
	return &UpdateFeeComponentUseCase{feeComponentRepo: feeComponentRepo}
}

func (uc *UpdateFeeComponentUseCase) Execute(ctx context.Context, id string, req dto.UpdateFeeComponentRequest) (*dto.FeeComponentResponse, error) {
	fc, err := uc.feeComponentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, feeConst.CodeFeeComponentNotFound)
	}

	var periodType *feeConst.PeriodType
	if req.PeriodType != nil {
		pt := feeConst.PeriodType(*req.PeriodType)
		periodType = &pt
	}

	fc.Update(req.Name, req.Amount, req.IsPeriodic, periodType, req.Description)

	if err := uc.feeComponentRepo.Update(ctx, fc); err != nil {
		return nil, application.WrapRepoErr(err, feeConst.CodeFeeComponentNotFound)
	}

	return toFeeComponentResponse(fc), nil
}
