package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	setConst "sipon-be/internal/modules/keuangan/domain/setting/constant"
	setEntity "sipon-be/internal/modules/keuangan/domain/setting/entity"
	setRepo "sipon-be/internal/modules/keuangan/domain/setting/repository"
	"sipon-be/internal/shared/kernel"
)

type UpdateKeuanganSettingUseCase struct {
	settingRepo setRepo.KeuanganSettingRepository
	accountRepo accRepo.AccountRepository
}

func NewUpdateKeuanganSettingUseCase(settingRepo setRepo.KeuanganSettingRepository, accountRepo accRepo.AccountRepository) *UpdateKeuanganSettingUseCase {
	return &UpdateKeuanganSettingUseCase{settingRepo: settingRepo, accountRepo: accountRepo}
}

func (uc *UpdateKeuanganSettingUseCase) Execute(ctx context.Context, req dto.UpdateKeuanganSettingRequest) (*dto.KeuanganSettingResponse, error) {
	if req.DefaultPaymentDebitAccountID != nil {
		acc, err := uc.accountRepo.FindByID(ctx, *req.DefaultPaymentDebitAccountID)
		if err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				switch ke.Code {
				case accConst.CodeAccountNotFound:
					return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
				}
			}
			return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
		if err := acc.EnsurePostable(); err != nil {
			return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "Akun default pembayaran harus postable dan aktif", err)
		}
		if acc.SubType == nil || *acc.SubType != accConst.SubTypeCashBank {
			return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "Akun default pembayaran harus merupakan akun kas atau bank", nil)
		}
	}

	setting, err := uc.settingRepo.Find(ctx)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case setConst.CodeSettingNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if err := setting.SetDefaultPaymentDebitAccountID(req.DefaultPaymentDebitAccountID); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "gagal menyimpan settings", err)
	}
	if err := uc.settingRepo.Update(ctx, setting); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	return uc.toResponse(ctx, setting)
}

func (uc *UpdateKeuanganSettingUseCase) toResponse(ctx context.Context, setting *setEntity.KeuanganSetting) (*dto.KeuanganSettingResponse, error) {
	accountID, err := setting.GetDefaultPaymentDebitAccountID()
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "gagal membaca settings", err)
	}
	resp := &dto.KeuanganSettingResponse{
		DefaultPaymentDebitAccountID: accountID,
	}
	if accountID != nil {
		acc, err := uc.accountRepo.FindByID(ctx, *accountID)
		if err == nil {
			resp.DefaultPaymentDebitAccount = &dto.AccountBriefResponse{
				ID:      acc.ID,
				Code:    acc.Code,
				Name:    acc.Name,
				Type:    string(acc.Type),
				SubType: subTypeStr(acc.SubType),
			}
		}
	}
	return resp, nil
}
