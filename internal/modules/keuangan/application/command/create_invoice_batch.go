package command

import (
	"context"
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
	bsEntity "sipon-be/internal/modules/keuangan/domain/billingscheme/entity"
	bsRepo "sipon-be/internal/modules/keuangan/domain/billingscheme/repository"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invEntity "sipon-be/internal/modules/keuangan/domain/invoice/entity"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
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
	}
}

type CreateInvoiceBatchCmd struct {
	BillingSchemeID string
	BillingPeriodID string
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
		return nil, application.WrapRepoErr(err, invConst.CodeInvoiceNotFound)
	}
	if !scheme.IsActive {
		return nil, kernel.New(invConst.CodeInvoiceInvalidStatus)
	}

	period, err := uc.billingPeriodRepo.FindByID(ctx, cmd.BillingPeriodID)
	if err != nil {
		return nil, application.WrapRepoErr(err, bpConst.CodeBillingPeriodNotFound)
	}
	if !period.IsOpen() {
		return nil, kernel.New(bpConst.CodeBillingPeriodInvalidStatus)
	}

	dueDate, err := time.Parse("2006-01-02", cmd.DueDate)
	if err != nil {
		return nil, application.WrapRepoErr(err, invConst.CodeInvoiceNotFound)
	}

	santriInfos, err := uc.kesantrianReader.ListActiveSantriWithUserID(ctx)
	if err != nil {
		return nil, application.WrapRepoErr(err, invConst.CodeInvoiceNotFound)
	}

	// Step 1: batch header, status processing.
	batch, err := bbEntity.NewBillingBatch(
		uuid.New().String(), fmt.Sprintf("%s - %s", scheme.Name, period.Name),
		cmd.BillingSchemeID, cmd.BillingPeriodID, cmd.CreatedBy,
	)
	if err != nil {
		return nil, application.WrapRepoErr(err, bbConst.CodeBillingBatchNotFound)
	}
	if err := uc.batchRepo.Save(ctx, batch); err != nil {
		return nil, application.WrapRepoErr(err, bbConst.CodeBillingBatchNotFound)
	}

	// Step 2: snapshot every active santri as a target, up front.
	targets := make([]*bbEntity.BillingBatchTarget, 0, len(santriInfos))
	eligible := make([]eligibleBatchTarget, 0, len(santriInfos))

	for _, info := range santriInfos {
		assignment, err := uc.assignmentRepo.FindActiveBySantriID(ctx, info.SantriID)
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
			return nil, application.WrapRepoErr(err, bbConst.CodeBillingBatchNotFound)
		}
	}

	// Step 3: process each eligible target in its own transaction.
	created, skipped, errored := 0, len(targets)-len(eligible), 0

	for _, et := range eligible {
		finalStatus, invoiceID, reason := uc.processTarget(ctx, et, scheme, cmd, dueDate)

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
			return nil, application.WrapRepoErr(err, bbConst.CodeBillingBatchNotFound)
		}
	}

	// Step 4: close out the batch header with totals.
	batch.Complete(created, skipped, errored)
	if err := uc.batchRepo.Update(ctx, batch); err != nil {
		return nil, application.WrapRepoErr(err, bbConst.CodeBillingBatchNotFound)
	}

	return &dto.CreateInvoiceBatchResponse{BatchID: batch.ID, Status: string(batch.Status)}, nil
}

func (uc *CreateInvoiceBatchUseCase) processTarget(
	ctx context.Context,
	et eligibleBatchTarget,
	scheme *bsEntity.BillingScheme,
	cmd CreateInvoiceBatchCmd,
	dueDate time.Time,
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
				uuid.New().String(), invNum, et.santriID, et.userID,
				item.FeeComponentID, cmd.BillingPeriodID,
				amount, dueDate, cmd.CreatedBy,
			)
			if err != nil {
				errMsgs = append(errMsgs, err.Error())
				continue
			}
			inv.BillingSchemeID = &cmd.BillingSchemeID
			if err := inv.Issue(); err != nil {
				errMsgs = append(errMsgs, err.Error())
				continue
			}
			if err := uc.invoiceRepo.Save(txCtx, inv); err != nil {
				errMsgs = append(errMsgs, err.Error())
				continue
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
