package query

import (
	"context"

	journalConst "sipon-be/internal/modules/keuangan/domain/journal/constant"
	journalRepo "sipon-be/internal/modules/keuangan/domain/journal/repository"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
)

type ListJournalEntriesUseCase struct {
	journalRepo journalRepo.JournalRepository
}

func NewListJournalEntriesUseCase(journalRepo journalRepo.JournalRepository) *ListJournalEntriesUseCase {
	return &ListJournalEntriesUseCase{journalRepo: journalRepo}
}

func (uc *ListJournalEntriesUseCase) Execute(ctx context.Context, query dto.JournalListQuery) ([]dto.JournalEntryResponse, *dto.Meta, error) {
	repoQuery := journalRepo.JournalListQuery{
		PeriodID:   query.PeriodID,
		Status:     query.Status,
		SourceType: query.SourceType,
		Page:       query.Page,
		Limit:      query.Limit,
	}
	if repoQuery.Page == 0 {
		repoQuery.Page = 1
	}
	if repoQuery.Limit == 0 {
		repoQuery.Limit = 20
	}

	result, err := uc.journalRepo.List(ctx, repoQuery)
	if err != nil {
		return nil, nil, application.WrapRepoErr(err, journalConst.CodeJournalQueryFailed)
	}

	items := make([]dto.JournalEntryResponse, len(result.Items))
	for i, entry := range result.Items {
		resp := dto.JournalEntryResponse{
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
		items[i] = resp
	}

	totalPages := (result.Total + int64(repoQuery.Limit) - 1) / int64(repoQuery.Limit)
	meta := &dto.Meta{
		Page:       repoQuery.Page,
		Limit:      repoQuery.Limit,
		Total:      result.Total,
		TotalPages: totalPages,
	}

	return items, meta, nil
}
