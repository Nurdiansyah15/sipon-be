package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	bsRepo "sipon-be/internal/modules/keuangan/domain/billingscheme/repository"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invEntity "sipon-be/internal/modules/keuangan/domain/invoice/entity"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	invVO "sipon-be/internal/modules/keuangan/domain/invoice/valueobject"
	"sipon-be/internal/shared/kernel"
)

type CreateInvoiceBatchUseCase struct {
	invoiceRepo      invRepo.InvoiceRepository
	feeRepo          feeRepo.FeeComponentRepository
	schemeRepo       bsRepo.BillingSchemeRepository
	assignmentRepo   bsRepo.SantriBillingAssignmentRepository
	kesantrianReader ports.KesantrianReader
	transactor       ports.Transactor
}

func NewCreateInvoiceBatchUseCase(
	invoiceRepo invRepo.InvoiceRepository,
	feeRepo feeRepo.FeeComponentRepository,
	schemeRepo bsRepo.BillingSchemeRepository,
	assignmentRepo bsRepo.SantriBillingAssignmentRepository,
	kesantrianReader ports.KesantrianReader,
	transactor ports.Transactor,
) *CreateInvoiceBatchUseCase {
	return &CreateInvoiceBatchUseCase{
		invoiceRepo:      invoiceRepo,
		feeRepo:          feeRepo,
		schemeRepo:       schemeRepo,
		assignmentRepo:   assignmentRepo,
		kesantrianReader: kesantrianReader,
		transactor:       transactor,
	}
}

type CreateInvoiceBatchCmd struct {
	BillingSchemeID string
	Periode         string
	TahunAjaran     string
	DueDate         string
	CreatedBy       string
}

func (uc *CreateInvoiceBatchUseCase) Execute(ctx context.Context, cmd CreateInvoiceBatchCmd) (*dto.MessageResponse, error) {
	scheme, err := uc.schemeRepo.FindByID(ctx, cmd.BillingSchemeID)
	if err != nil {
		return nil, application.WrapRepoErr(err, invConst.CodeInvoiceNotFound)
	}
	if !scheme.IsActive {
		return nil, kernel.New(invConst.CodeInvoiceInvalidStatus)
	}

	santriIDs, err := uc.kesantrianReader.ListActiveSantriIDs(ctx)
	if err != nil {
		return nil, application.WrapRepoErr(err, invConst.CodeInvoiceNotFound)
	}

	dueDate, err := time.Parse("2006-01-02", cmd.DueDate)
	if err != nil {
		return nil, application.WrapRepoErr(err, invConst.CodeInvoiceNotFound)
	}

	created := 0
	skipped := 0
	errors := make([]string, 0)

	for _, santriID := range santriIDs {
		assignment, err := uc.assignmentRepo.FindActiveBySantriID(ctx, santriID)
		if err != nil || assignment.BillingSchemeID != cmd.BillingSchemeID {
			skipped++
			continue
		}

		userID := assignment.SantriID

		for _, item := range scheme.Items {
			fee, err := uc.feeRepo.FindByID(ctx, item.FeeComponentID)
			if err != nil {
				skipped++
				continue
			}

			existing, _ := uc.invoiceRepo.FindBySantriComponentPeriode(ctx, santriID, item.FeeComponentID, cmd.Periode)
			if existing != nil {
				skipped++
				continue
			}

			amount := item.GetEffectiveAmount(fee.Amount)
			invNum := invVO.NewInvoiceNumberNow(created + 1)
			inv, err := invEntity.NewInvoice(
				uuid.New().String(), invNum.String(), santriID, userID,
				item.FeeComponentID, cmd.Periode, cmd.TahunAjaran,
				amount, dueDate, cmd.CreatedBy,
			)
			if err != nil {
				errors = append(errors, fmt.Sprintf("santri %s: %v", santriID, err))
				skipped++
				continue
			}
			inv.BillingSchemeID = &cmd.BillingSchemeID
			if err := inv.Issue(); err != nil {
				skipped++
				continue
			}
			if err := uc.invoiceRepo.Save(ctx, inv); err != nil {
				skipped++
				continue
			}
			created++
		}
	}

	msg := fmt.Sprintf("Batch invoice selesai: %d dibuat, %d dilewati", created, skipped)
	if len(errors) > 0 {
		msg += fmt.Sprintf(", %d error", len(errors))
	}
	return &dto.MessageResponse{Message: msg}, nil
}
