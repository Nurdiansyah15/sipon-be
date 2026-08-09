package query

import (
	"context"
	"errors"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	journalConst "sipon-be/internal/modules/keuangan/domain/journal/constant"
	journalEntity "sipon-be/internal/modules/keuangan/domain/journal/entity"
	journalRepo "sipon-be/internal/modules/keuangan/domain/journal/repository"
	"sipon-be/internal/shared/kernel"
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
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case journalConst.CodeJournalNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	lines, err := uc.journalRepo.FindLinesByEntryID(ctx, entry.ID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	return journalEntryToResponse(entry, lines), nil
}

func journalEntryToResponse(entry *journalEntity.JournalEntry, lines []*journalEntity.JournalEntryLine) *dto.JournalEntryResponse {
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

	return resp
}
