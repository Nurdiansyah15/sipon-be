package command

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	journalConst "sipon-be/internal/modules/keuangan/domain/journal/constant"
	journalEntity "sipon-be/internal/modules/keuangan/domain/journal/entity"
	journalRepo "sipon-be/internal/modules/keuangan/domain/journal/repository"
	periodConst "sipon-be/internal/modules/keuangan/domain/period/constant"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/shared/kernel"
)

type CreateManualJournalUseCase struct {
	journalRepo journalRepo.JournalRepository
	accountRepo accRepo.AccountRepository
	periodRepo  periodRepo.AccountingPeriodRepository
	transactor  ports.Transactor
}

func NewCreateManualJournalUseCase(journalRepo journalRepo.JournalRepository, accountRepo accRepo.AccountRepository, periodRepo periodRepo.AccountingPeriodRepository, transactor ports.Transactor) *CreateManualJournalUseCase {
	return &CreateManualJournalUseCase{journalRepo: journalRepo, accountRepo: accountRepo, periodRepo: periodRepo, transactor: transactor}
}

func (uc *CreateManualJournalUseCase) Execute(ctx context.Context, req dto.CreateJournalEntryRequest, postedBy string) (*dto.JournalEntryResponse, error) {
	period, err := uc.periodRepo.FindActive(ctx)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case periodConst.CodePeriodNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if req.PeriodID != "" {
		period, err = uc.periodRepo.FindByID(ctx, req.PeriodID)
		if err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				switch ke.Code {
				case periodConst.CodePeriodNotFound:
					return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
				}
			}
			return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
	}
	if !period.CanPost() {
		return nil, kernel.WrapMsg(application.ErrCodeConflict, "Periode akuntansi sudah ditutup", nil)
	}

	entryDate, err := time.Parse("2006-01-02", req.EntryDate)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "format tanggal tidak valid", err)
	}

	var savedEntry *journalEntity.JournalEntry
	err = uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		jn, err := uc.journalRepo.NextJournalNumber(txCtx)
		if err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				switch ke.Code {
				case journalConst.CodeJournalNotFound:
					return kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
				}
			}
			return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
		entry, err := journalEntity.NewJournalEntry(
			uuid.New().String(), jn.String(), entryDate,
			req.Description, period.ID, postedBy,
		)
		if err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				switch ke.Code {
				case journalConst.CodeJournalNotFound:
					return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
				}
			}
			return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
		entry.SetSource(journalConst.SourceManual, entry.ID)

		for _, line := range req.Lines {
			acc, err := uc.accountRepo.FindByID(txCtx, line.AccountID)
			if err != nil {
				var ke *kernel.AppError
				if errors.As(err, &ke) {
					switch ke.Code {
					case accConst.CodeAccountNotFound:
						return kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
					}
				}
				return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
			}
			if err := acc.EnsurePostable(); err != nil {
				var ke *kernel.AppError
				if errors.As(err, &ke) {
					switch ke.Code {
					case accConst.CodeAccountNotPostable:
						return kernel.WrapMsg(application.ErrCodeBadRequest, ke.Message, ke)
					}
				}
				return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
			}

			entryLine := journalEntity.NewJournalEntryLine(
				uuid.New().String(), entry.ID,
				acc.ID, acc.Code,
				line.Debit, line.Credit, line.Description,
			)
			entry.AddLine(entryLine)
		}

		if err := entry.Post(); err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				switch ke.Code {
				case journalConst.CodeJournalNotBalanced,
					journalConst.CodeJournalMinLines:
					return kernel.WrapMsg(application.ErrCodeBadRequest, ke.Message, ke)
				case journalConst.CodeJournalInvalidStatus:
					return kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
				}
			}
			return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}

		if err := uc.journalRepo.Save(txCtx, entry); err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				switch ke.Code {
				case journalConst.CodeJournalNotFound:
					return kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
				}
			}
			return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
		savedEntry = entry
		return nil
	})
	if err != nil {
		return nil, err
	}

	return toJournalEntryResponse(savedEntry), nil
}

func toJournalEntryResponse(entry *journalEntity.JournalEntry) *dto.JournalEntryResponse {
	resp := &dto.JournalEntryResponse{
		ID:            entry.ID,
		JournalNumber: entry.JournalNumber,
		EntryDate:     entry.EntryDate.Format("2006-01-02"),
		Description:   entry.Description,
		PeriodID:      entry.PeriodID,
		TotalDebit:    entry.TotalDebit,
		TotalCredit:   entry.TotalCredit,
		Status:        string(entry.Status),
		PostedBy:      entry.PostedBy,
		CreatedAt:     entry.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     entry.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if entry.SourceType != nil {
		s := string(*entry.SourceType)
		resp.SourceType = &s
	}
	if entry.SourceID != nil {
		resp.SourceID = entry.SourceID
	}
	if entry.PostedAt != nil {
		s := entry.PostedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.PostedAt = &s
	}
	if entry.Lines != nil {
		resp.Lines = make([]dto.JournalLineResponse, len(entry.Lines))
		for i, l := range entry.Lines {
			resp.Lines[i] = dto.JournalLineResponse{
				ID:          l.ID,
				AccountID:   l.AccountID,
				AccountCode: l.AccountCode,
				AccountName: l.AccountName,
				Description: l.Description,
				Debit:       l.Debit,
				Credit:      l.Credit,
			}
		}
	}
	return resp
}
