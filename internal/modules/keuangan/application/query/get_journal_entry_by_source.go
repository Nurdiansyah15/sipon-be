package query

import (
	"context"
	"errors"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	journalConst "sipon-be/internal/modules/keuangan/domain/journal/constant"
	journalRepo "sipon-be/internal/modules/keuangan/domain/journal/repository"
	"sipon-be/internal/shared/kernel"
)

type GetJournalEntryBySourceUseCase struct {
	journalRepo journalRepo.JournalRepository
}

func NewGetJournalEntryBySourceUseCase(journalRepo journalRepo.JournalRepository) *GetJournalEntryBySourceUseCase {
	return &GetJournalEntryBySourceUseCase{journalRepo: journalRepo}
}

func (uc *GetJournalEntryBySourceUseCase) Execute(ctx context.Context, sourceType, sourceID string) (*dto.JournalEntryResponse, error) {
	entry, err := uc.journalRepo.FindBySource(ctx, sourceType, sourceID)
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
	return journalEntryToResponse(entry, entry.Lines), nil
}
