package query

import (
	"context"
	"errors"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	setConst "sipon-be/internal/modules/keuangan/domain/setting/constant"
	setEntity "sipon-be/internal/modules/keuangan/domain/setting/entity"
	setRepo "sipon-be/internal/modules/keuangan/domain/setting/repository"
	"sipon-be/internal/shared/kernel"
)

type GetKeuanganSettingUseCase struct {
	settingRepo setRepo.KeuanganSettingRepository
	accountRepo accRepo.AccountRepository
}

func NewGetKeuanganSettingUseCase(settingRepo setRepo.KeuanganSettingRepository, accountRepo accRepo.AccountRepository) *GetKeuanganSettingUseCase {
	return &GetKeuanganSettingUseCase{settingRepo: settingRepo, accountRepo: accountRepo}
}

func (uc *GetKeuanganSettingUseCase) Execute(ctx context.Context) (*dto.KeuanganSettingResponse, error) {
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
	return uc.toResponse(ctx, setting)
}

func (uc *GetKeuanganSettingUseCase) toResponse(ctx context.Context, setting *setEntity.KeuanganSetting) (*dto.KeuanganSettingResponse, error) {
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
