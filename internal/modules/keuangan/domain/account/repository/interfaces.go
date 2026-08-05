package repository

import (
	"context"

	"sipon-be/internal/modules/keuangan/domain/account/entity"
)

type AccountListQuery struct {
	Type   *string
	Active *bool
	Page   int
	Limit  int
}

type AccountListResult struct {
	Items []*entity.Account
	Total int64
}

type AccountRepository interface {
	Save(ctx context.Context, acc *entity.Account) error
	Update(ctx context.Context, acc *entity.Account) error
	FindByID(ctx context.Context, id string) (*entity.Account, error)
	FindByCode(ctx context.Context, code string) (*entity.Account, error)
	List(ctx context.Context, query AccountListQuery) (*AccountListResult, error)
	ListAll(ctx context.Context) ([]*entity.Account, error)
	ListPostable(ctx context.Context) ([]*entity.Account, error)
	FindChildren(ctx context.Context, parentID string) ([]*entity.Account, error)
	HasJournalEntries(ctx context.Context, accountID string) (bool, error)
	ExistsByCode(ctx context.Context, code string, excludeID string) (bool, error)
}
