package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	feeConst "sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	journalConst "sipon-be/internal/modules/keuangan/domain/journal/constant"
	journalService "sipon-be/internal/modules/keuangan/domain/journal/service"
	periodConst "sipon-be/internal/modules/keuangan/domain/period/constant"
	"sipon-be/internal/shared/kernel"
)

type CancelInvoiceUseCase struct {
	invoiceRepo invRepo.InvoiceRepository
	feeRepo     feeRepo.FeeComponentRepository
	transactor  ports.Transactor
	autoPosting *journalService.AutoPostingService
}

func NewCancelInvoiceUseCase(invoiceRepo invRepo.InvoiceRepository, feeRepo feeRepo.FeeComponentRepository, transactor ports.Transactor, autoPosting *journalService.AutoPostingService) *CancelInvoiceUseCase {
	return &CancelInvoiceUseCase{invoiceRepo: invoiceRepo, feeRepo: feeRepo, transactor: transactor, autoPosting: autoPosting}
}

func (uc *CancelInvoiceUseCase) Execute(ctx context.Context, id string, cancelledBy string) (*dto.InvoiceResponse, error) {
	var resp *dto.InvoiceResponse
	err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		inv, err := uc.invoiceRepo.FindByID(txCtx, id)
		if err != nil {
			return err
		}

		if err := inv.Cancel(); err != nil {
			return err
		}

		if err := uc.invoiceRepo.Update(txCtx, inv); err != nil {
			return err
		}

		if inv.IssuedAt != nil && uc.autoPosting != nil {
			fee, err := uc.feeRepo.FindByID(txCtx, inv.FeeComponentID)
			if err != nil {
				return err
			}
			if err := uc.autoPosting.PostInvoiceCancelled(
				txCtx, inv.ID, inv.InvoiceNumber, "",
				*inv.IssuedAt, inv.Amount, fee.RevenueAccountID, fee.ReceivableAccountID, cancelledBy,
			); err != nil {
				return err
			}
		}

		resp = toInvoiceResponse(inv, nil)
		return nil
	})
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case invConst.CodeInvoiceNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			case invConst.CodeInvoiceInvalidStatus:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			case feeConst.CodeFeeComponentNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			case journalConst.CodeJournalAccountMappingNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			case journalConst.CodeJournalPeriodClosed:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			case accConst.CodeAccountNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			case periodConst.CodePeriodNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	return resp, nil
}
