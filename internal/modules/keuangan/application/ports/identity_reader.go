package ports

import "context"

type IdentityReader interface {
	GetUserSummary(ctx context.Context, userID string) (string, string, error)
}
