package query

import (
	"context"

	journalConst "sipon-be/internal/modules/keuangan/domain/journal/constant"
	journalRepo "sipon-be/internal/modules/keuangan/domain/journal/repository"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
)

type GetJournalEntryUseCase struct {
	journalRepo journalRepo.JournalRepository
}

func NewGetJournalEntryUseCase(journalRepo journalRepo.JournalRepository) *GetJournalEntryUseCase {
	return &GetJournalEntryUseCase{journalRepo: journalRepo}
}

func (uc *GetJournalEntryUseCase) Execute(ctx context.Context, id string) (*dto.JournalEntryResponse, error) {
	entry, err := uc.journalRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, journalConst.CodeJournalNotFound)
	}

	lines, err := uc.journalRepo.FindLinesByEntryID(ctx, entry.ID)
	if err != nil {
		return nil, application.WrapRepoErr(err, journalConst.CodeJournalNotFound)
	}

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
	if lines != nil {
		resp.Lines = make([]dto.JournalLineResponse, len(lines))
		for i, l := range lines {
			resp.Lines[i] = dto.JournalLineResponse{
				ID:          l.ID,
				AccountID:   l.AccountID,
				AccountCode: l.AccountCode,
				Description: l.Description,
				Debit:       l.Debit,
				Credit:      l.Credit,
			}
		}
	}

	return resp, nil
}
