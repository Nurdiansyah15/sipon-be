package command

import (
	"context"

	journalConst "sipon-be/internal/modules/keuangan/domain/journal/constant"
	journalRepo "sipon-be/internal/modules/keuangan/domain/journal/repository"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
)

type CancelJournalUseCase struct {
	journalRepo journalRepo.JournalRepository
}

func NewCancelJournalUseCase(journalRepo journalRepo.JournalRepository) *CancelJournalUseCase {
	return &CancelJournalUseCase{journalRepo: journalRepo}
}

func (uc *CancelJournalUseCase) Execute(ctx context.Context, journalID string) (*dto.MessageResponse, error) {
	entry, err := uc.journalRepo.FindByID(ctx, journalID)
	if err != nil {
		return nil, application.WrapRepoErr(err, journalConst.CodeJournalNotFound)
	}

	if err := entry.Cancel(); err != nil {
		return nil, application.WrapRepoErr(err, journalConst.CodeJournalInvalidStatus)
	}

	if err := uc.journalRepo.Update(ctx, entry); err != nil {
		return nil, application.WrapRepoErr(err, journalConst.CodeJournalNotFound)
	}

	return &dto.MessageResponse{Message: "Jurnal berhasil dibatalkan"}, nil
}
