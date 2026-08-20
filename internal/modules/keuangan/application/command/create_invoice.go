package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	santriConst "sipon-be/internal/modules/kesantrian/domain/santri/constant"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	bpConst "sipon-be/internal/modules/keuangan/domain/billingperiod/constant"
	bpEntity "sipon-be/internal/modules/keuangan/domain/billingperiod/entity"
	bpRepo "sipon-be/internal/modules/keuangan/domain/billingperiod/repository"
	billRepo "sipon-be/internal/modules/keuangan/domain/billingscheme/repository"
	feeConst "sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invEntity "sipon-be/internal/modules/keuangan/domain/invoice/entity"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	journalConst "sipon-be/internal/modules/keuangan/domain/journal/constant"
	journalService "sipon-be/internal/modules/keuangan/domain/journal/service"
	periodConst "sipon-be/internal/modules/keuangan/domain/period/constant"
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
	outboxWriter      ports.OutboxWriter
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

func (uc *CreateInvoiceUseCase) SetOutboxWriter(w ports.OutboxWriter) {
	uc.outboxWriter = w
}

func (uc *CreateInvoiceUseCase) publishInvoiceIssued(ctx context.Context, userID, invoiceID, invoiceNumber string) {
	if uc.outboxWriter == nil {
		return
	}
	payload, _ := json.Marshal(map[string]string{
		"user_id":        userID,
		"invoice_id":     invoiceID,
		"invoice_number": invoiceNumber,
	})
	if err := uc.outboxWriter.Save(ctx, RoutingInvoiceIssued, payload); err != nil {
		slog.Warn("keuangan: gagal publish event", "routing_key", RoutingInvoiceIssued, "invoice_id", invoiceID, "error", err)
	}
}

type CreateInvoiceCmd struct {
	SantriID        string
	FeeComponentID  string
	BillingSchemeID *string
	BillingPeriodID *string
	IssuedDate      string
	Amount          float64
	DueDate         string
	Notes           *string
	CreatedBy       string
	Issue           bool
}

func (uc *CreateInvoiceUseCase) Execute(ctx context.Context, cmd CreateInvoiceCmd) (*dto.InvoiceResponse, error) {
	fee, err := uc.feeRepo.FindByID(ctx, cmd.FeeComponentID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case feeConst.CodeFeeComponentNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if !fee.IsActive {
		return nil, kernel.WrapMsg(application.ErrCodeNotFound, "Komponen biaya tidak aktif", nil)
	}

	issuedDate, err := time.Parse("2006-01-02", cmd.IssuedDate)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "format tanggal terbit tidak valid", err)
	}

	var billingPeriod *bpEntity.BillingPeriod
	if fee.IsPeriodic && (cmd.BillingPeriodID == nil || *cmd.BillingPeriodID == "") {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "Komponen biaya ini periodik, periode tagihan wajib diisi", nil)
	}
	if cmd.BillingPeriodID != nil && *cmd.BillingPeriodID != "" {
		billingPeriod, err = uc.billingPeriodRepo.FindByID(ctx, *cmd.BillingPeriodID)
		if err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) && ke.Code == bpConst.CodeBillingPeriodNotFound {
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
			return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
		if !billingPeriod.IsOpen() {
			return nil, kernel.WrapMsg(application.ErrCodeConflict, "Status periode tagihan tidak valid", nil)
		}
		if issuedDate.Before(billingPeriod.StartDate) || issuedDate.After(billingPeriod.EndDate) {
			return nil, kernel.WrapMsg(application.ErrCodeBadRequest,
				fmt.Sprintf("Tanggal terbit harus dalam rentang periode tagihan %s (%s s.d. %s)",
					billingPeriod.Name, billingPeriod.StartDate.Format("2006-01-02"), billingPeriod.EndDate.Format("2006-01-02")), nil)
		}
	}

	santri, err := uc.kesantrianReader.GetSantriByID(ctx, cmd.SantriID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case santriConst.CodeSantriNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if billingPeriod != nil {
		existing, _ := uc.invoiceRepo.FindBySantriComponentPeriod(ctx, cmd.SantriID, cmd.FeeComponentID, billingPeriod.ID)
		if existing != nil {
			return nil, kernel.WrapMsg(application.ErrCodeConflict, "Invoice duplikat", nil)
		}
	}

	dueDate, err := time.Parse("2006-01-02", cmd.DueDate)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "format tanggal tidak valid", err)
	}

	invNum, err := uc.invoiceRepo.NextInvoiceNumber(ctx)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	inv, err := invEntity.NewInvoice(
		uuid.New().String(), invNum.String(), cmd.SantriID, santri.UserID,
		cmd.FeeComponentID, cmd.BillingPeriodID,
		cmd.Amount, dueDate, cmd.CreatedBy,
	)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case invConst.CodeInvoiceNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	inv.BillingSchemeID = cmd.BillingSchemeID
	inv.Notes = cmd.Notes

	if cmd.Issue {
		if err := inv.Issue(issuedDate); err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				switch ke.Code {
				case invConst.CodeInvoiceInvalidStatus:
					return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
				}
			}
			return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
	}

	err = uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := uc.invoiceRepo.Save(txCtx, inv); err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				switch ke.Code {
				case invConst.CodeInvoiceDuplicate:
					return kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
				}
			}
			return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
		if cmd.Issue && uc.autoPosting != nil && inv.IssuedAt != nil {
			if err := uc.autoPosting.PostInvoiceIssued(
				txCtx, inv.ID, inv.InvoiceNumber, "",
				*inv.IssuedAt, inv.Amount, inv.DiscountAmount, fee.RevenueAccountID, fee.ReceivableAccountID, cmd.CreatedBy,
			); err != nil {
				var ke *kernel.AppError
				if errors.As(err, &ke) {
					switch ke.Code {
					case journalConst.CodeJournalAccountMappingNotFound:
						return kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
					case journalConst.CodeJournalPeriodClosed:
						return kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
					case accConst.CodeAccountNotFound:
						return kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
					case periodConst.CodePeriodNotFound:
						return kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
					}
				}
				return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if cmd.Issue {
		uc.publishInvoiceIssued(ctx, inv.UserID, inv.ID, inv.InvoiceNumber)
	}

	return toInvoiceResponse(inv, billingPeriod), nil
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
