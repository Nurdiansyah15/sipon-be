package command

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	bbConst "sipon-be/internal/modules/keuangan/domain/billingbatch/constant"
	bbEntity "sipon-be/internal/modules/keuangan/domain/billingbatch/entity"
	bbRepo "sipon-be/internal/modules/keuangan/domain/billingbatch/repository"
	bpConst "sipon-be/internal/modules/keuangan/domain/billingperiod/constant"
	bpRepo "sipon-be/internal/modules/keuangan/domain/billingperiod/repository"
	bsConst "sipon-be/internal/modules/keuangan/domain/billingscheme/constant"
	bsEntity "sipon-be/internal/modules/keuangan/domain/billingscheme/entity"
	bsRepo "sipon-be/internal/modules/keuangan/domain/billingscheme/repository"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	invEntity "sipon-be/internal/modules/keuangan/domain/invoice/entity"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	journalService "sipon-be/internal/modules/keuangan/domain/journal/service"
	"sipon-be/internal/shared/kernel"
)

type CreateInvoiceBatchUseCase struct {
	invoiceRepo       invRepo.InvoiceRepository
	feeRepo           feeRepo.FeeComponentRepository
	schemeRepo        bsRepo.BillingSchemeRepository
	assignmentRepo    bsRepo.SantriBillingAssignmentRepository
	billingPeriodRepo bpRepo.BillingPeriodRepository
	batchRepo         bbRepo.BillingBatchRepository
	targetRepo        bbRepo.BillingBatchTargetRepository
	kesantrianReader  ports.KesantrianReader
	transactor        ports.Transactor
	autoPosting       *journalService.AutoPostingService
}

func NewCreateInvoiceBatchUseCase(
	invoiceRepo invRepo.InvoiceRepository,
	feeRepo feeRepo.FeeComponentRepository,
	schemeRepo bsRepo.BillingSchemeRepository,
	assignmentRepo bsRepo.SantriBillingAssignmentRepository,
	billingPeriodRepo bpRepo.BillingPeriodRepository,
	batchRepo bbRepo.BillingBatchRepository,
	targetRepo bbRepo.BillingBatchTargetRepository,
	kesantrianReader ports.KesantrianReader,
	transactor ports.Transactor,
	autoPosting *journalService.AutoPostingService,
) *CreateInvoiceBatchUseCase {
	return &CreateInvoiceBatchUseCase{
		invoiceRepo:       invoiceRepo,
		feeRepo:           feeRepo,
		schemeRepo:        schemeRepo,
		assignmentRepo:    assignmentRepo,
		billingPeriodRepo: billingPeriodRepo,
		batchRepo:         batchRepo,
		targetRepo:        targetRepo,
		kesantrianReader:  kesantrianReader,
		transactor:        transactor,
		autoPosting:       autoPosting,
	}
}

type CreateInvoiceBatchCmd struct {
	BillingSchemeID string
	BillingPeriodID string
	IssuedDate      string
	DueDate         string
	CreatedBy       string
}

type eligibleBatchTarget struct {
	santriID string
	userID   string
	target   *bbEntity.BillingBatchTarget
}

func (uc *CreateInvoiceBatchUseCase) Execute(ctx context.Context, cmd CreateInvoiceBatchCmd) (*dto.CreateInvoiceBatchResponse, error) {
	scheme, err := uc.schemeRepo.FindByID(ctx, cmd.BillingSchemeID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case bsConst.CodeBillingSchemeNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if !scheme.IsActive {
		return nil, kernel.WrapMsg(application.ErrCodeConflict, "Skema tagihan tidak aktif", nil)
	}

	period, err := uc.billingPeriodRepo.FindByID(ctx, cmd.BillingPeriodID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case bpConst.CodeBillingPeriodNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if !period.IsOpen() {
		return nil, kernel.WrapMsg(application.ErrCodeConflict, "Status periode tagihan tidak valid", nil)
	}

	issuedDate, err := time.Parse("2006-01-02", cmd.IssuedDate)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "format tanggal terbit tidak valid", err)
	}
	if issuedDate.Before(period.StartDate) || issuedDate.After(period.EndDate) {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest,
			fmt.Sprintf("Tanggal terbit harus dalam rentang periode tagihan %s (%s s.d. %s)",
				period.Name, period.StartDate.Format("2006-01-02"), period.EndDate.Format("2006-01-02")), nil)
	}

	dueDate, err := time.Parse("2006-01-02", cmd.DueDate)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "format tanggal tidak valid", err)
	}

	santriInfos, err := uc.kesantrianReader.ListActiveSantriWithUserID(ctx)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	// Step 1: batch header, status processing.
	batch, err := bbEntity.NewBillingBatch(
		uuid.New().String(), fmt.Sprintf("%s - %s", scheme.Name, period.Name),
		cmd.BillingSchemeID, cmd.BillingPeriodID, cmd.CreatedBy,
	)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case bbConst.CodeBillingBatchNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if err := uc.batchRepo.Save(ctx, batch); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	// Step 2: snapshot every active santri as a target, up front.
	targets := make([]*bbEntity.BillingBatchTarget, 0, len(santriInfos))
	eligible := make([]eligibleBatchTarget, 0, len(santriInfos))

	for _, info := range santriInfos {
		assignment, err := uc.assignmentRepo.FindActiveBySantriIDAt(ctx, info.SantriID, issuedDate)
		var status bbConst.BillingBatchTargetStatus
		switch {
		case err != nil:
			status = bbConst.TargetSkippedNoAssignment
		case assignment.BillingSchemeID != cmd.BillingSchemeID:
			status = bbConst.TargetSkippedWrongScheme
		default:
			status = bbConst.TargetPending
		}

		target := bbEntity.NewBillingBatchTarget(uuid.New().String(), batch.ID, info.SantriID, status)
		targets = append(targets, target)
		if status == bbConst.TargetPending {
			eligible = append(eligible, eligibleBatchTarget{santriID: info.SantriID, userID: info.UserID, target: target})
		}
	}
	if len(targets) > 0 {
		if err := uc.targetRepo.SaveMany(ctx, targets); err != nil {
			return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
	}

	// Step 3: process each eligible target in its own transaction.
	created, skipped, errored := 0, len(targets)-len(eligible), 0

	for _, et := range eligible {
		finalStatus, invoiceID, reason := uc.processTarget(ctx, et, scheme, cmd, dueDate, issuedDate)

		switch finalStatus {
		case bbConst.TargetCreated:
			created++
		case bbConst.TargetError:
			errored++
		default:
			skipped++
		}

		et.target.MarkProcessed(finalStatus, invoiceID, reason)
		if err := uc.targetRepo.UpdateTarget(ctx, et.target); err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				switch ke.Code {
				case bbConst.CodeBillingBatchNotFound:
					return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
				}
			}
			return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
	}

	// Step 4: close out the batch header with totals.
	batch.Complete(created, skipped, errored)
	if err := uc.batchRepo.Update(ctx, batch); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case bbConst.CodeBillingBatchNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	return &dto.CreateInvoiceBatchResponse{BatchID: batch.ID, Status: string(batch.Status)}, nil
}

func (uc *CreateInvoiceBatchUseCase) processTarget(
	ctx context.Context,
	et eligibleBatchTarget,
	scheme *bsEntity.BillingScheme,
	cmd CreateInvoiceBatchCmd,
	dueDate time.Time,
	issuedDate time.Time,
) (bbConst.BillingBatchTargetStatus, *string, *string) {
	var (
		finalStatus bbConst.BillingBatchTargetStatus
		invoiceID   *string
		reason      *string
	)

	txErr := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		createdCount, alreadyCount, missingCount := 0, 0, 0
		var errMsgs []string
		var lastInvoiceID string

		for _, item := range scheme.Items {
			fee, err := uc.feeRepo.FindByID(txCtx, item.FeeComponentID)
			if err != nil {
				missingCount++
				continue
			}

			existing, _ := uc.invoiceRepo.FindBySantriComponentPeriod(txCtx, et.santriID, item.FeeComponentID, cmd.BillingPeriodID)
			if existing != nil {
				alreadyCount++
				continue
			}

			amount := item.GetEffectiveAmount(fee.Amount)
			invNum, err := uc.invoiceRepo.NextInvoiceNumber(txCtx)
			if err != nil {
				errMsgs = append(errMsgs, err.Error())
				continue
			}
			inv, err := invEntity.NewInvoice(
				uuid.New().String(), invNum.String(), et.santriID, et.userID,
				item.FeeComponentID, &cmd.BillingPeriodID,
				amount, dueDate, cmd.CreatedBy,
			)
			if err != nil {
				errMsgs = append(errMsgs, err.Error())
				continue
			}
			inv.BillingSchemeID = &cmd.BillingSchemeID
			if err := inv.Issue(issuedDate); err != nil {
				errMsgs = append(errMsgs, err.Error())
				continue
			}
			if err := uc.invoiceRepo.Save(txCtx, inv); err != nil {
				errMsgs = append(errMsgs, err.Error())
				continue
			}
			if uc.autoPosting != nil && inv.IssuedAt != nil {
				if err := uc.autoPosting.PostInvoiceIssued(
					txCtx, inv.ID, inv.InvoiceNumber, "",
					*inv.IssuedAt, inv.Amount, inv.DiscountAmount, fee.Type, cmd.CreatedBy,
				); err != nil {
					return err
				}
			}
			createdCount++
			lastInvoiceID = inv.ID
		}

		switch {
		case createdCount > 0:
			finalStatus = bbConst.TargetCreated
			invoiceID = &lastInvoiceID
			if alreadyCount+missingCount+len(errMsgs) > 0 {
				r := fmt.Sprintf("%d komponen dibuat, %d sudah ditagih, %d bermasalah", createdCount, alreadyCount, missingCount+len(errMsgs))
				reason = &r
			}
		case len(errMsgs) > 0:
			finalStatus = bbConst.TargetError
			r := strings.Join(errMsgs, "; ")
			reason = &r
		case missingCount > 0 && alreadyCount == 0:
			finalStatus = bbConst.TargetSkippedComponentMissing
		default:
			finalStatus = bbConst.TargetSkippedAlreadyInvoiced
		}
		return nil
	})
	if txErr != nil {
		finalStatus = bbConst.TargetError
		msg := txErr.Error()
		reason = &msg
	}

	return finalStatus, invoiceID, reason
}
