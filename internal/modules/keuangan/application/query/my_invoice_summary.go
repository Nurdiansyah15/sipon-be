package query

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	"sipon-be/internal/shared/kernel"
)

type MyInvoiceSummaryUseCase struct {
	invoiceRepo invRepo.InvoiceRepository
}

func NewMyInvoiceSummaryUseCase(invoiceRepo invRepo.InvoiceRepository) *MyInvoiceSummaryUseCase {
	return &MyInvoiceSummaryUseCase{invoiceRepo: invoiceRepo}
}

func (uc *MyInvoiceSummaryUseCase) Execute(ctx context.Context, userID string) (*dto.MyInvoiceSummaryResponse, error) {
	s, err := uc.invoiceRepo.FindSummaryByUserID(ctx, userID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	return &dto.MyInvoiceSummaryResponse{
		TotalTagihan:   s.TotalTagihan,
		TotalTerbayar:  s.TotalTerbayar,
		TotalTunggakan: s.TotalTunggakan,
		JumlahInvoice:  s.JumlahInvoice,
		JumlahLunas:    s.JumlahLunas,
		JumlahBelum:    s.JumlahBelum,
	}, nil
}
