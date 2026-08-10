package command

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	accEntity "sipon-be/internal/modules/keuangan/domain/account/entity"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	feeConst "sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	feeEntity "sipon-be/internal/modules/keuangan/domain/feecomponent/entity"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	"sipon-be/internal/shared/kernel"
)

type CreateFeeComponentUseCase struct {
	feeComponentRepo feeRepo.FeeComponentRepository
	accountRepo      accRepo.AccountRepository
}

func NewCreateFeeComponentUseCase(feeComponentRepo feeRepo.FeeComponentRepository, accountRepo accRepo.AccountRepository) *CreateFeeComponentUseCase {
	return &CreateFeeComponentUseCase{feeComponentRepo: feeComponentRepo, accountRepo: accountRepo}
}

func (uc *CreateFeeComponentUseCase) Execute(ctx context.Context, req dto.CreateFeeComponentRequest, createdBy string) (*dto.FeeComponentResponse, error) {
	exists, err := uc.feeComponentRepo.ExistsByCode(ctx, req.Code, "")
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if exists {
		return nil, kernel.WrapMsg(application.ErrCodeConflict, "Komponen biaya dengan kode yang sama sudah ada", nil)
	}

	revenue, err := resolveFeeRevenueAccount(ctx, uc.accountRepo, req.RevenueAccountID)
	if err != nil {
		return nil, err
	}
	receivable, err := resolveFeeReceivableAccount(ctx, uc.accountRepo, req.ReceivableAccountID)
	if err != nil {
		return nil, err
	}

	var periodType *feeConst.PeriodType
	if req.PeriodType != nil {
		pt := feeConst.PeriodType(*req.PeriodType)
		periodType = &pt
	}

	fc, err := feeEntity.NewFeeComponent(uuid.New().String(), req.Code, req.Name, req.RevenueAccountID, req.ReceivableAccountID, req.Amount, createdBy)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case feeConst.CodeFeeComponentNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
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

	return toFeeComponentResponse(fc, revenue, receivable), nil
}

// resolveFeeRevenueAccount memvalidasi akun pendapatan yang dipilih untuk
// komponen biaya: harus bertipe revenue, postable, dan aktif.
func resolveFeeRevenueAccount(ctx context.Context, repo accRepo.AccountRepository, accountID string) (*accEntity.Account, error) {
	acc, err := repo.FindByID(ctx, accountID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == accConst.CodeAccountNotFound {
			return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "Akun pendapatan tidak ditemukan", ke)
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if acc.Type != accConst.TypeRevenue {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "Akun pendapatan harus merupakan akun bertipe revenue", nil)
	}
	if err := acc.EnsurePostable(); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "Akun pendapatan harus postable dan aktif", err)
	}
	return acc, nil
}

// resolveFeeReceivableAccount memvalidasi akun piutang yang dipilih untuk
// komponen biaya: harus bertipe asset dengan sub-tipe receivable, postable,
// dan aktif.
func resolveFeeReceivableAccount(ctx context.Context, repo accRepo.AccountRepository, accountID string) (*accEntity.Account, error) {
	acc, err := repo.FindByID(ctx, accountID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == accConst.CodeAccountNotFound {
			return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "Akun piutang tidak ditemukan", ke)
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if acc.SubType == nil || *acc.SubType != accConst.SubTypeReceivable {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "Akun piutang harus merupakan akun dengan sub-tipe receivable", nil)
	}
	if err := acc.EnsurePostable(); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "Akun piutang harus postable dan aktif", err)
	}
	return acc, nil
}

func toFeeComponentResponse(fc *feeEntity.FeeComponent, revenue, receivable *accEntity.Account) *dto.FeeComponentResponse {
	resp := &dto.FeeComponentResponse{
		ID:         fc.ID,
		Code:       fc.Code,
		Name:       fc.Name,
		RevenueAccount: &dto.AccountBriefResponse{
			ID:      revenue.ID,
			Code:    revenue.Code,
			Name:    revenue.Name,
			Type:    string(revenue.Type),
			SubType: subTypeStr(revenue.SubType),
		},
		ReceivableAccount: &dto.AccountBriefResponse{
			ID:      receivable.ID,
			Code:    receivable.Code,
			Name:    receivable.Name,
			Type:    string(receivable.Type),
			SubType: subTypeStr(receivable.SubType),
		},
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
