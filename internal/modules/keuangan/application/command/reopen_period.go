package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	journalConst "sipon-be/internal/modules/keuangan/domain/journal/constant"
	journalRepo "sipon-be/internal/modules/keuangan/domain/journal/repository"
	periodConst "sipon-be/internal/modules/keuangan/domain/period/constant"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/shared/kernel"
)

type ReopenPeriodUseCase struct {
	periodRepo  periodRepo.AccountingPeriodRepository
	journalRepo journalRepo.JournalRepository
	transactor  ports.Transactor
}

func NewReopenPeriodUseCase(periodRepo periodRepo.AccountingPeriodRepository, journalRepo journalRepo.JournalRepository, transactor ports.Transactor) *ReopenPeriodUseCase {
	return &ReopenPeriodUseCase{periodRepo: periodRepo, journalRepo: journalRepo, transactor: transactor}
}

func (uc *ReopenPeriodUseCase) Execute(ctx context.Context, periodID string) (*dto.PeriodResponse, error) {
	var resp *dto.PeriodResponse
	err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		period, err := uc.periodRepo.FindByID(txCtx, periodID)
		if err != nil {
			return err
		}

		if err := period.Reopen(); err != nil {
			return err
		}

		if closingEntry, err := uc.journalRepo.FindBySource(txCtx, string(journalConst.SourceClosing), periodID); err == nil {
			if err := closingEntry.Cancel(); err != nil {
				return err
			}
			if err := uc.journalRepo.Update(txCtx, closingEntry); err != nil {
				return err
			}
		} else {
			var ke *kernel.AppError
			if !errors.As(err, &ke) || ke.Code != journalConst.CodeJournalNotFound {
				return err
			}
		}

		if err := uc.periodRepo.Update(txCtx, period); err != nil {
			return err
		}
		resp = toPeriodResponse(period)
		return nil
	})
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case periodConst.CodePeriodNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			case periodConst.CodePeriodInvalidStatus:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			case journalConst.CodeJournalAutoCannotCancel:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			case journalConst.CodeJournalInvalidStatus:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	return resp, nil
}
