package entity

import (
	"time"

	"sipon-be/internal/modules/identity/domain/user/constant"
	"sipon-be/internal/modules/identity/domain/user/valueobject"
	"sipon-be/internal/shared/kernel"
)

type User struct {
	ID                  string
	Username            valueobject.Username
	Fullname            *string
	Email               valueobject.Email
	PhoneNumber         *valueobject.PhoneNumber
	AvatarKey           *string
	Status              constant.UserStatus
	Credentials         []*Credential
	CreatedAt           time.Time
	UpdatedAt           time.Time
	LastLoginAt         *time.Time
	DeletedAt           *time.Time
	FailedLoginAttempts int
	LockedUntil         *time.Time
}

func NewUser(id string, username valueobject.Username, fullname *string, email valueobject.Email, phoneNumber *valueobject.PhoneNumber) (*User, error) {
	if id == "" {
		return nil, kernel.WrapMsg(constant.ErrCodeUserIDRequired, "ID pengguna wajib diisi", nil)
	}
	if email == (valueobject.Email{}) {
		return nil, kernel.WrapMsg(constant.ErrCodeUserEmailRequired, "Email pengguna wajib diisi", nil)
	}
	if phoneNumber != nil && *phoneNumber == (valueobject.PhoneNumber{}) {
		return nil, kernel.WrapMsg(constant.ErrCodeUserPhoneNumberInvalid, "Nomor telepon pengguna tidak valid", nil)
	}

	now := time.Now()
	return &User{
		ID:          id,
		Username:    username,
		Fullname:    fullname,
		Email:       email,
		PhoneNumber: phoneNumber,
		Status:      constant.UserStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (u *User) AddCredential(c *Credential) {
	u.Credentials = append(u.Credentials, c)
	u.UpdatedAt = time.Now()
}

func (u *User) FindLoginIdentity(kind constant.LoginIdentifierKind, value string) *LoginIdentity {
	for _, c := range u.Credentials {
		if li := c.FindLoginIdentity(kind, value); li != nil {
			return li
		}
	}
	return nil
}

func (u *User) FindLoginIdentityByKind(kind constant.LoginIdentifierKind) *LoginIdentity {
	for _, c := range u.Credentials {
		if li := c.FindLoginIdentityByKind(kind); li != nil {
			return li
		}
	}
	return nil
}

func (u *User) FindCredential(typ constant.CredentialType) *Credential {
	for _, c := range u.Credentials {
		if c.Type == typ {
			return c
		}
	}
	return nil
}

func (u *User) EnsureCanLogin() error {
	if u.Status == constant.UserStatusBanned {
		return kernel.WrapMsg(constant.ErrCodeUserBanned, "Pengguna telah diblokir", nil)
	}
	if u.DeletedAt != nil {
		return kernel.WrapMsg(constant.ErrCodeUserNotActive, "Pengguna tidak aktif", nil)
	}
	return u.EnsureNotLockedOut()
}

func (u *User) Activate() error {
	if u.Status == constant.UserStatusActive {
		return kernel.WrapMsg(constant.ErrCodeUserAlreadyActive, "Pengguna sudah aktif", nil)
	}
	u.Status = constant.UserStatusActive
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) Deactivate() error {
	if u.Status == constant.UserStatusBanned {
		return kernel.WrapMsg(constant.ErrCodeUserAlreadyBanned, "Pengguna sudah diblokir", nil)
	}
	u.Status = constant.UserStatusBanned
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) Reactivate() error {
	if u.Status != constant.UserStatusBanned {
		return kernel.WrapMsg(constant.ErrCodeUserNotActive, "Pengguna tidak aktif", nil)
	}
	u.Status = constant.UserStatusActive
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) MarkLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

func (u *User) IsLockedOut() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

func (u *User) EnsureNotLockedOut() error {
	if u.IsLockedOut() {
		return kernel.WrapMsg(constant.ErrCodeUserLockedOut, "Pengguna terkunci sementara", nil)
	}
	return nil
}

func (u *User) IncrementFailedAttempts() {
	u.FailedLoginAttempts++
	if u.FailedLoginAttempts >= constant.MaxLoginAttempts {
		lockUntil := time.Now().Add(constant.LockoutDuration)
		u.LockedUntil = &lockUntil
	}
	u.UpdatedAt = time.Now()
}

func (u *User) ResetFailedAttempts() {
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
	u.UpdatedAt = time.Now()
}

func (u *User) SoftDelete() error {
	if u.DeletedAt != nil {
		return kernel.WrapMsg(constant.ErrCodeUserAlreadyDeleted, "Pengguna sudah dihapus", nil)
	}
	now := time.Now()
	u.DeletedAt = &now
	u.UpdatedAt = now
	return nil
}

func (u *User) HasLocalPassword() bool {
	cred := u.FindCredential(constant.CredentialTypeLocal)
	if cred == nil {
		return false
	}
	return cred.SecretHash != nil
}

func (u *User) SetLocalPassword(hashed valueobject.HashedPassword) error {
	cred := u.FindCredential(constant.CredentialTypeLocal)
	if cred == nil {
		return kernel.WrapMsg(constant.ErrCodeCredentialNotLocal, "Kredensial bukan tipe lokal", nil)
	}
	cred.SecretHash = &hashed
	now := time.Now()
	cred.LastChangedAt = &now
	cred.UpdatedAt = now
	u.UpdatedAt = now
	return nil
}

func (u *User) ChangeUsername(newUsername valueobject.Username) {
	u.Username = newUsername
	u.UpdatedAt = time.Now()
}
