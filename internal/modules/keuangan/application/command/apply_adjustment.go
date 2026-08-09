package command

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	adjConst "sipon-be/internal/modules/keuangan/domain/adjustment/constant"
	adjEntity "sipon-be/internal/modules/keuangan/domain/adjustment/entity"
	adjRepo "sipon-be/internal/modules/keuangan/domain/adjustment/repository"
	feeConst "sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	journalConst "sipon-be/internal/modules/keuangan/domain/journal/constant"
	journalService "sipon-be/internal/modules/keuangan/domain/journal/service"
	periodConst "sipon-be/internal/modules/keuangan/domain/period/constant"
	"sipon-be/internal/shared/kernel"
)

type ApplyAdjustmentUseCase struct {
	adjRepo     adjRepo.AdjustmentRepository
	invoiceRepo invRepo.InvoiceRepository
	feeRepo     feeRepo.FeeComponentRepository
	transactor  ports.Transactor
	autoPosting *journalService.AutoPostingService
}

func NewApplyAdjustmentUseCase(adjRepo adjRepo.AdjustmentRepository, invoiceRepo invRepo.InvoiceRepository, feeRepo feeRepo.FeeComponentRepository, transactor ports.Transactor, autoPosting *journalService.AutoPostingService) *ApplyAdjustmentUseCase {
	return &ApplyAdjustmentUseCase{adjRepo: adjRepo, invoiceRepo: invoiceRepo, feeRepo: feeRepo, transactor: transactor, autoPosting: autoPosting}
}

func (uc *ApplyAdjustmentUseCase) Execute(ctx context.Context, invoiceID string, req dto.ApplyAdjustmentRequest, appliedBy string) (*dto.InvoiceResponse, error) {
	var resp *dto.InvoiceResponse
	err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		inv, err := uc.invoiceRepo.FindByID(txCtx, invoiceID)
		if err != nil {
			return err
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
			return err
		}

		if err := uc.adjRepo.Save(txCtx, adj); err != nil {
			return err
		}

		inv.ApplyDiscount(adjustmentAmount)
		if err := uc.invoiceRepo.Update(txCtx, inv); err != nil {
			return err
		}

		if inv.IssuedAt != nil && adjustmentAmount > 0 && uc.autoPosting != nil {
			fee, err := uc.feeRepo.FindByID(txCtx, inv.FeeComponentID)
			if err != nil {
				return err
			}
			if err := uc.autoPosting.PostAdjustment(
				txCtx, adj.ID, inv.InvoiceNumber, "",
				*inv.IssuedAt, adjustmentAmount, fee.Type, appliedBy,
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
			case feeConst.CodeFeeComponentNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			case journalConst.CodeJournalAccountMappingNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			case journalConst.CodeJournalPeriodClosed:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			case periodConst.CodePeriodNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	return resp, nil
}
