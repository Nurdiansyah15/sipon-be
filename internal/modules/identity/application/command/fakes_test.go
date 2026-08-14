package command

import (
	"context"
	"time"

	ports "sipon-be/internal/modules/identity/application/ports"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userentity "sipon-be/internal/modules/identity/domain/user/entity"
	uservo "sipon-be/internal/modules/identity/domain/user/valueobject"
	"sipon-be/internal/shared/kernel"
)

type fakeGoogleVerifier struct {
	info *ports.GoogleIdentityInfo
	err  error
}

func (f *fakeGoogleVerifier) VerifyIDToken(ctx context.Context, idToken string, allowedClientIDs []string) (*ports.GoogleIdentityInfo, error) {
	return f.info, f.err
}

type fakeTokenGen struct {
	access  string
	refresh string
}

func (f *fakeTokenGen) GenerateAccessToken(userID, sessionID, deviceID string) (string, error) {
	return f.access, nil
}
func (f *fakeTokenGen) GenerateRefreshToken(userID, deviceID string) (string, error) {
	return f.refresh, nil
}
func (f *fakeTokenGen) ParseAccessToken(token string) (*ports.TokenClaims, error) {
	return nil, nil
}
func (f *fakeTokenGen) ParseRefreshToken(token string) (*ports.RefreshTokenClaims, error) {
	return nil, nil
}

type fakeTransactor struct{}

func (f *fakeTransactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func nowPtr() *time.Time {
	t := time.Now()
	return &t
}

type fakeUserRepo struct {
	users      map[string]*userentity.User
	byIdentity map[string]string // "kind|value" -> userID
	updated    []*userentity.User
	saved      []*userentity.User
	errSave    error
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		users:      map[string]*userentity.User{},
		byIdentity: map[string]string{},
	}
}

func (f *fakeUserRepo) index(user *userentity.User) {
	f.users[user.ID] = user
	for _, cred := range user.Credentials {
		for _, li := range cred.LoginIdentities {
			f.byIdentity[string(li.Kind)+"|"+li.Value] = user.ID
		}
	}
}

func (f *fakeUserRepo) Save(ctx context.Context, user *userentity.User) error {
	if f.errSave != nil {
		return f.errSave
	}
	f.index(user)
	f.saved = append(f.saved, user)
	return nil
}

func (f *fakeUserRepo) Update(ctx context.Context, user *userentity.User) error {
	f.index(user)
	f.updated = append(f.updated, user)
	return nil
}

func (f *fakeUserRepo) FindByID(ctx context.Context, id string) (*userentity.User, error) {
	if u, ok := f.users[id]; ok {
		return u, nil
	}
	return nil, kernel.WrapMsg(userconstant.ErrCodeUserNotFound, "Pengguna tidak ditemukan", nil)
}

func (f *fakeUserRepo) FindByLoginIdentifier(ctx context.Context, identifier uservo.LoginIdentifier) (*userentity.User, error) {
	return f.FindByIdentity(ctx, identifier.Kind, identifier.Value)
}

func (f *fakeUserRepo) FindByIdentity(ctx context.Context, kind userconstant.LoginIdentifierKind, value string) (*userentity.User, error) {
	if uid, ok := f.byIdentity[string(kind)+"|"+value]; ok {
		if u, ok2 := f.users[uid]; ok2 {
			return u, nil
		}
	}
	return nil, kernel.WrapMsg(userconstant.ErrCodeUserNotFound, "Pengguna tidak ditemukan", nil)
}

func (f *fakeUserRepo) FindByUsername(ctx context.Context, username string) (*userentity.User, error) {
	for _, u := range f.users {
		if u.Username.String() == username {
			return u, nil
		}
	}
	return nil, kernel.WrapMsg(userconstant.ErrCodeUserNotFound, "Pengguna tidak ditemukan", nil)
}

func (f *fakeUserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	_, err := f.FindByUsername(ctx, username)
	if err != nil {
		var ke *kernel.AppError
		if ok := asKernel(err, &ke); ok && ke.Code == userconstant.ErrCodeUserNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (f *fakeUserRepo) ExistsByLoginIdentity(ctx context.Context, kind userconstant.LoginIdentifierKind, value string) (bool, error) {
	_, ok := f.byIdentity[string(kind)+"|"+value]
	return ok, nil
}

func (f *fakeUserRepo) UpdateUsername(ctx context.Context, userID, newUsername string) error {
	if u, ok := f.users[userID]; ok {
		un, err := uservo.NewUsername(newUsername)
		if err != nil {
			return err
		}
		u.ChangeUsername(un)
	}
	return nil
}

func asKernel(err error, target **kernel.AppError) bool {
	ke, ok := err.(*kernel.AppError)
	if ok {
		*target = ke
	}
	return ok
}
