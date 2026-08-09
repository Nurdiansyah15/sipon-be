package command

import (
	"context"
	"errors"

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
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if exists {
		return nil, kernel.WrapMsg(application.ErrCodeConflict, "Komponen biaya dengan kode yang sama sudah ada", nil)
	}

	feeType := feeConst.FeeComponentType(req.Type)
	if !feeConst.IsValidFeeType(feeType) {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "Jenis komponen biaya tidak valid", nil)
	}

	var periodType *feeConst.PeriodType
	if req.PeriodType != nil {
		pt := feeConst.PeriodType(*req.PeriodType)
		periodType = &pt
	}

	fc, err := feeEntity.NewFeeComponent(uuid.New().String(), req.Code, req.Name, feeType, req.Amount, createdBy)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case feeConst.CodeFeeComponentNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
			case feeConst.CodeFeeComponentInvalidType:
				return nil, kernel.WrapMsg(application.ErrCodeBadRequest, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	fc.IsPeriodic = req.IsPeriodic
	fc.PeriodType = periodType
	fc.Description = req.Description

	if err := uc.feeComponentRepo.Save(ctx, fc); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case feeConst.CodeFeeComponentDuplicate:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
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
