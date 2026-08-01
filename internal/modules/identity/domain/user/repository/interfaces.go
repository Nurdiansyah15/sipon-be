package repository

import (
	"context"

	"sipon-be/internal/modules/identity/domain/user/constant"
	"sipon-be/internal/modules/identity/domain/user/entity"
	"sipon-be/internal/modules/identity/domain/user/valueobject"
)

type UserRepository interface {
	Save(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByLoginIdentifier(ctx context.Context, identifier valueobject.LoginIdentifier) (*entity.User, error)
	FindByIdentity(ctx context.Context, kind constant.LoginIdentifierKind, value string) (*entity.User, error)
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByLoginIdentity(ctx context.Context, kind constant.LoginIdentifierKind, value string) (bool, error)
	UpdateUsername(ctx context.Context, userID, newUsername string) error
}
