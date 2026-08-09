package repository

import (
	"context"

	"sipon-be/internal/modules/keuangan/domain/journal/entity"
	"sipon-be/internal/modules/keuangan/domain/journal/valueobject"
)

type JournalListQuery struct {
	PeriodID   *string
	Status     *string
	SourceType *string
	Page       int
	Limit      int
}

type JournalListResult struct {
	Items []*entity.JournalEntry
	Total int64
}

type AccountBalance struct {
	Debit  float64
	Credit float64
}

type JournalRepository interface {
	Save(ctx context.Context, entry *entity.JournalEntry) error
	Update(ctx context.Context, entry *entity.JournalEntry) error
	FindByID(ctx context.Context, id string) (*entity.JournalEntry, error)
	FindByNumber(ctx context.Context, number string) (*entity.JournalEntry, error)
	NextJournalNumber(ctx context.Context) (valueobject.JournalNumber, error)
	List(ctx context.Context, query JournalListQuery) (*JournalListResult, error)
	FindBySource(ctx context.Context, sourceType string, sourceID string) (*entity.JournalEntry, error)
	SaveLines(ctx context.Context, entryID string, lines []*entity.JournalEntryLine) error
	FindLinesByEntryID(ctx context.Context, entryID string) ([]*entity.JournalEntryLine, error)
	ComputeAccountBalances(ctx context.Context, periodID string) (map[string]AccountBalance, error)
}
