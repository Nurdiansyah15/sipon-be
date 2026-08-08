package command

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	journalService "sipon-be/internal/modules/keuangan/domain/journal/service"
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
				*inv.IssuedAt, inv.Amount, fee.Type, cancelledBy,
			); err != nil {
				return err
			}
		}

		resp = toInvoiceResponse(inv, nil)
		return nil
	})
	if err != nil {
		return nil, application.WrapRepoErr(err, invConst.CodeInvoiceNotFound)
	}

	return resp, nil
}
