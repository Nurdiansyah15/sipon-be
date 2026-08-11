package ports

import "context"

type Transactor interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
