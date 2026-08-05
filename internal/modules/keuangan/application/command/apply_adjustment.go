package command

import (
	"context"

	"github.com/google/uuid"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	adjConst "sipon-be/internal/modules/keuangan/domain/adjustment/constant"
	adjEntity "sipon-be/internal/modules/keuangan/domain/adjustment/entity"
	adjRepo "sipon-be/internal/modules/keuangan/domain/adjustment/repository"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
)

type ApplyAdjustmentUseCase struct {
	adjRepo     adjRepo.AdjustmentRepository
	invoiceRepo invRepo.InvoiceRepository
}

func NewApplyAdjustmentUseCase(adjRepo adjRepo.AdjustmentRepository, invoiceRepo invRepo.InvoiceRepository) *ApplyAdjustmentUseCase {
	return &ApplyAdjustmentUseCase{adjRepo: adjRepo, invoiceRepo: invoiceRepo}
}

func (uc *ApplyAdjustmentUseCase) Execute(ctx context.Context, invoiceID string, req dto.ApplyAdjustmentRequest, appliedBy string) (*dto.InvoiceResponse, error) {
	inv, err := uc.invoiceRepo.FindByID(ctx, invoiceID)
	if err != nil {
		return nil, application.WrapRepoErr(err, invConst.CodeInvoiceNotFound)
	}

	adjustmentAmount := req.Amount
	if req.Percentage != nil {
		adjustmentAmount = inv.Amount * (*req.Percentage) / 100.0
	}

	adjType := adjConst.AdjustmentType(req.Type)
	adj, err := adjEntity.NewInvoiceAdjustment(
		uuid.New().String(), invoiceID, adjType,
		adjustmentAmount, req.Percentage, req.Description, appliedBy,
	)
	if err != nil {
		return nil, application.WrapRepoErr(err, adjConst.CodeAdjustmentNotFound)
	}

	if err := uc.adjRepo.Save(ctx, adj); err != nil {
		return nil, application.WrapRepoErr(err, adjConst.CodeAdjustmentNotFound)
	}

	inv.ApplyDiscount(adjustmentAmount)
	if err := uc.invoiceRepo.Update(ctx, inv); err != nil {
		return nil, application.WrapRepoErr(err, invConst.CodeInvoiceNotFound)
	}

	return toInvoiceResponse(inv), nil
}
