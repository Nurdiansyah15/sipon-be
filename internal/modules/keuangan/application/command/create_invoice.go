package command

import (
	"context"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	billRepo "sipon-be/internal/modules/keuangan/domain/billingscheme/repository"
	feeConst "sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invEntity "sipon-be/internal/modules/keuangan/domain/invoice/entity"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	invVO "sipon-be/internal/modules/keuangan/domain/invoice/valueobject"
	"sipon-be/internal/shared/kernel"
)

type CreateInvoiceUseCase struct {
	invoiceRepo    invRepo.InvoiceRepository
	feeRepo        feeRepo.FeeComponentRepository
	assignmentRepo billRepo.SantriBillingAssignmentRepository
	transactor     ports.Transactor
}

func NewCreateInvoiceUseCase(invoiceRepo invRepo.InvoiceRepository, feeRepo feeRepo.FeeComponentRepository, assignmentRepo billRepo.SantriBillingAssignmentRepository, transactor ports.Transactor) *CreateInvoiceUseCase {
	return &CreateInvoiceUseCase{
		invoiceRepo:    invoiceRepo,
		feeRepo:        feeRepo,
		assignmentRepo: assignmentRepo,
		transactor:     transactor,
	}
}

type CreateInvoiceCmd struct {
	SantriID       string
	UserID         string
	FeeComponentID string
	BillingSchemeID *string
	Periode        string
	TahunAjaran    string
	Amount         float64
	DueDate        string
	Notes          *string
	CreatedBy      string
	Issue          bool
}

func (uc *CreateInvoiceUseCase) Execute(ctx context.Context, cmd CreateInvoiceCmd) (*dto.InvoiceResponse, error) {
	fee, err := uc.feeRepo.FindByID(ctx, cmd.FeeComponentID)
	if err != nil {
		return nil, application.WrapRepoErr(err, feeConst.CodeFeeComponentNotFound)
	}
	if !fee.IsActive {
		return nil, kernel.New(feeConst.CodeFeeComponentNotFound)
	}

	existing, _ := uc.invoiceRepo.FindBySantriComponentPeriode(ctx, cmd.SantriID, cmd.FeeComponentID, cmd.Periode)
	if existing != nil {
		return nil, kernel.New(invConst.CodeInvoiceDuplicate)
	}

	dueDate, err := time.Parse("2006-01-02", cmd.DueDate)
	if err != nil {
		return nil, application.WrapRepoErr(err, invConst.CodeInvoiceNotFound)
	}

	invNum := invVO.NewInvoiceNumberNow(1)
	inv, err := invEntity.NewInvoice(
		uuid.New().String(), invNum.String(), cmd.SantriID, cmd.UserID,
		cmd.FeeComponentID, cmd.Periode, cmd.TahunAjaran,
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

	if err := uc.invoiceRepo.Save(ctx, inv); err != nil {
		return nil, application.WrapRepoErr(err, invConst.CodeInvoiceNotFound)
	}

	return toInvoiceResponse(inv), nil
}

func toInvoiceResponse(inv *invEntity.Invoice) *dto.InvoiceResponse {
	resp := &dto.InvoiceResponse{
		ID:             inv.ID,
		InvoiceNumber:  inv.InvoiceNumber,
		SantriID:       inv.SantriID,
		UserID:         inv.UserID,
		BillingSchemeID: inv.BillingSchemeID,
		FeeComponentID: inv.FeeComponentID,
		Periode:        inv.Periode,
		TahunAjaran:    inv.TahunAjaran,
		Amount:         inv.Amount,
		DiscountAmount: inv.DiscountAmount,
		PaidAmount:     inv.PaidAmount,
		Status:         string(inv.Status),
		DueDate:        inv.DueDate.Format("2006-01-02"),
		Notes:          inv.Notes,
		CreatedAt:      inv.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      inv.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if inv.IssuedAt != nil {
		s := inv.IssuedAt.Format("2006-01-02")
		resp.IssuedAt = &s
	}
	return resp
}
