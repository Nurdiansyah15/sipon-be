package command

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
)

type CancelInvoiceUseCase struct {
	invoiceRepo invRepo.InvoiceRepository
}

func NewCancelInvoiceUseCase(invoiceRepo invRepo.InvoiceRepository) *CancelInvoiceUseCase {
	return &CancelInvoiceUseCase{invoiceRepo: invoiceRepo}
}

func (uc *CancelInvoiceUseCase) Execute(ctx context.Context, id string) (*dto.InvoiceResponse, error) {
	inv, err := uc.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, invConst.CodeInvoiceNotFound)
	}

	if err := inv.Cancel(); err != nil {
		return nil, application.WrapRepoErr(err, invConst.CodeInvoiceInvalidStatus)
	}

	if err := uc.invoiceRepo.Update(ctx, inv); err != nil {
		return nil, application.WrapRepoErr(err, invConst.CodeInvoiceNotFound)
	}

	return toInvoiceResponse(inv, nil), nil
}
