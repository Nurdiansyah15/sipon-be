package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	bpConst "sipon-be/internal/modules/keuangan/domain/billingperiod/constant"
	bpEntity "sipon-be/internal/modules/keuangan/domain/billingperiod/entity"
	bpRepo "sipon-be/internal/modules/keuangan/domain/billingperiod/repository"
	billRepo "sipon-be/internal/modules/keuangan/domain/billingscheme/repository"
	feeConst "sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	invConst 	"sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invEntity "sipon-be/internal/modules/keuangan/domain/invoice/entity"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	journalConst "sipon-be/internal/modules/keuangan/domain/journal/constant"
	journalService "sipon-be/internal/modules/keuangan/domain/journal/service"
	"sipon-be/internal/shared/kernel"
)

type CreateInvoiceUseCase struct {
	invoiceRepo       invRepo.InvoiceRepository
	feeRepo           feeRepo.FeeComponentRepository
	assignmentRepo    billRepo.SantriBillingAssignmentRepository
	billingPeriodRepo bpRepo.BillingPeriodRepository
	kesantrianReader  ports.KesantrianReader
	transactor        ports.Transactor
	autoPosting       *journalService.AutoPostingService
}

func NewCreateInvoiceUseCase(invoiceRepo invRepo.InvoiceRepository, feeRepo feeRepo.FeeComponentRepository, assignmentRepo billRepo.SantriBillingAssignmentRepository, billingPeriodRepo bpRepo.BillingPeriodRepository, kesantrianReader ports.KesantrianReader, transactor ports.Transactor, autoPosting *journalService.AutoPostingService) *CreateInvoiceUseCase {
	return &CreateInvoiceUseCase{
		invoiceRepo:       invoiceRepo,
		feeRepo:           feeRepo,
		assignmentRepo:    assignmentRepo,
		billingPeriodRepo: billingPeriodRepo,
		kesantrianReader:  kesantrianReader,
		transactor:        transactor,
		autoPosting:       autoPosting,
	}
}

type CreateInvoiceCmd struct {
	SantriID        string
	FeeComponentID  string
	BillingSchemeID *string
	BillingPeriodID string
	Amount          float64
	DueDate         string
	Notes           *string
	CreatedBy       string
	Issue           bool
}

func (uc *CreateInvoiceUseCase) Execute(ctx context.Context, cmd CreateInvoiceCmd) (*dto.InvoiceResponse, error) {
	fee, err := uc.feeRepo.FindByID(ctx, cmd.FeeComponentID)
	if err != nil {
		return nil, application.WrapRepoErr(err, feeConst.CodeFeeComponentNotFound)
	}
	if !fee.IsActive {
		return nil, kernel.New(feeConst.CodeFeeComponentNotFound)
	}

	period, err := uc.billingPeriodRepo.FindByID(ctx, cmd.BillingPeriodID)
	if err != nil {
		return nil, application.WrapRepoErr(err, bpConst.CodeBillingPeriodNotFound)
	}
	if !period.IsOpen() {
		return nil, kernel.New(bpConst.CodeBillingPeriodInvalidStatus)
	}

	santri, err := uc.kesantrianReader.GetSantriByID(ctx, cmd.SantriID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, fmt.Errorf("santri not found: %w", err))
	}

	existing, _ := uc.invoiceRepo.FindBySantriComponentPeriod(ctx, cmd.SantriID, cmd.FeeComponentID, cmd.BillingPeriodID)
	if existing != nil {
		return nil, application.WrapConflictErr(kernel.New(invConst.CodeInvoiceDuplicate), invConst.CodeInvoiceDuplicate)
	}

	dueDate, err := time.Parse("2006-01-02", cmd.DueDate)
	if err != nil {
		return nil, application.WrapRepoErr(err, invConst.CodeInvoiceNotFound)
	}

	invNum, err := uc.invoiceRepo.NextInvoiceNumber(ctx)
	if err != nil {
		return nil, application.WrapRepoErr(err, invConst.CodeInvoicePersistenceFailed)
	}
	inv, err := invEntity.NewInvoice(
		uuid.New().String(), invNum.String(), cmd.SantriID, santri.UserID,
		cmd.FeeComponentID, cmd.BillingPeriodID,
		cmd.Amount, dueDate, cmd.CreatedBy,
	)
	if err != nil {
		return nil, application.WrapRepoErr(err, invConst.CodeInvoiceNotFound)
	}
	inv.BillingSchemeID = cmd.BillingSchemeID
	inv.Notes = cmd.Notes

	if cmd.Issue {
		if err := inv.Issue(); err != nil {
			return nil, application.WrapRepoErr(err, invConst.CodeInvoiceInvalidStatus)
		}
	}

	err = uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := uc.invoiceRepo.Save(txCtx, inv); err != nil {
			return application.WrapConflictErr(err, invConst.CodeInvoiceDuplicate)
		}
		if cmd.Issue && uc.autoPosting != nil && inv.IssuedAt != nil {
			if err := uc.autoPosting.PostInvoiceIssued(
				txCtx, inv.ID, inv.InvoiceNumber, "",
				*inv.IssuedAt, inv.Amount, inv.DiscountAmount, fee.Type, cmd.CreatedBy,
			); err != nil {
				return application.WrapConflictErr(err, journalConst.CodeJournalAccountMappingNotFound)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return toInvoiceResponse(inv, period), nil
}

func toInvoiceResponse(inv *invEntity.Invoice, period *bpEntity.BillingPeriod) *dto.InvoiceResponse {
	resp := &dto.InvoiceResponse{
		ID:              inv.ID,
		InvoiceNumber:   inv.InvoiceNumber,
		SantriID:        inv.SantriID,
		UserID:          inv.UserID,
		BillingSchemeID: inv.BillingSchemeID,
		FeeComponentID:  inv.FeeComponentID,
		BillingPeriodID: inv.BillingPeriodID,
		Amount:          inv.Amount,
		DiscountAmount:  inv.DiscountAmount,
		PaidAmount:      inv.PaidAmount,
		Status:          string(inv.Status),
		DueDate:         inv.DueDate.Format("2006-01-02"),
		Notes:           inv.Notes,
		CreatedAt:       inv.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       inv.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if inv.IssuedAt != nil {
		s := inv.IssuedAt.Format("2006-01-02")
		resp.IssuedAt = &s
	}
	if period != nil {
		resp.BillingPeriod = &dto.BillingPeriodBriefResponse{
			ID:     period.ID,
			Name:   period.Name,
			Status: string(period.Status),
		}
	}
	return resp
}
