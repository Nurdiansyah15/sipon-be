package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	journalConst "sipon-be/internal/modules/keuangan/domain/journal/constant"
	journalRepo "sipon-be/internal/modules/keuangan/domain/journal/repository"
	periodConst "sipon-be/internal/modules/keuangan/domain/period/constant"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/shared/kernel"
)

type CancelJournalUseCase struct {
	journalRepo journalRepo.JournalRepository
	periodRepo  periodRepo.AccountingPeriodRepository
}

func NewCancelJournalUseCase(journalRepo journalRepo.JournalRepository, periodRepo periodRepo.AccountingPeriodRepository) *CancelJournalUseCase {
	return &CancelJournalUseCase{journalRepo: journalRepo, periodRepo: periodRepo}
}

func (uc *CancelJournalUseCase) Execute(ctx context.Context, journalID string) (*dto.MessageResponse, error) {
	entry, err := uc.journalRepo.FindByID(ctx, journalID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case journalConst.CodeJournalNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	period, err := uc.periodRepo.FindByID(ctx, entry.PeriodID)
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
	if !period.IsOpen() {
		return nil, kernel.WrapMsg(application.ErrCodeConflict, "Periode akuntansi sudah ditutup atau dikunci, jurnal tidak dapat dibatalkan", nil)
	}

	if err := entry.Cancel(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case journalConst.CodeJournalInvalidStatus,
				journalConst.CodeJournalAutoCannotCancel:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if err := uc.journalRepo.Update(ctx, entry); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case journalConst.CodeJournalNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	return &dto.MessageResponse{Message: "Jurnal berhasil dibatalkan"}, nil
}
