package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	feeConst "sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	"sipon-be/internal/shared/kernel"
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
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case feeConst.CodeFeeComponentNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	var periodType *feeConst.PeriodType
	if req.PeriodType != nil {
		pt := feeConst.PeriodType(*req.PeriodType)
		periodType = &pt
	}

	fc.Update(req.Name, req.Amount, req.IsPeriodic, periodType, req.Description)

	if err := uc.feeComponentRepo.Update(ctx, fc); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case feeConst.CodeFeeComponentNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	return toFeeComponentResponse(fc), nil
}
