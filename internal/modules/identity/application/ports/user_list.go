package ports

import (
	"context"

	userentity "sipon-be/internal/modules/identity/domain/user/entity"
)

type UserListRepository interface {
	List(ctx context.Context, status string, roleID string, search string, sortBy string, sortType string, page, limit int) ([]*userentity.User, int64, error)
	FindByIDWithRoles(ctx context.Context, userID string) (*userentity.User, []string, error)
}
