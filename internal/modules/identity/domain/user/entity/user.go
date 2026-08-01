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
		return nil, kernel.New(constant.ErrCodeUserIDRequired)
	}
	if email == (valueobject.Email{}) {
		return nil, kernel.New(constant.ErrCodeUserEmailRequired)
	}
	if phoneNumber != nil && *phoneNumber == (valueobject.PhoneNumber{}) {
		return nil, kernel.New(constant.ErrCodeUserPhoneNumberInvalid)
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
		return kernel.New(constant.ErrCodeUserBanned)
	}
	if u.DeletedAt != nil {
		return kernel.New(constant.ErrCodeUserNotActive)
	}
	return u.EnsureNotLockedOut()
}

func (u *User) Activate() error {
	if u.Status == constant.UserStatusActive {
		return kernel.New(constant.ErrCodeUserAlreadyActive)
	}
	u.Status = constant.UserStatusActive
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) Deactivate() error {
	if u.Status == constant.UserStatusBanned {
		return kernel.New(constant.ErrCodeUserAlreadyBanned)
	}
	u.Status = constant.UserStatusBanned
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) Reactivate() error {
	if u.Status != constant.UserStatusBanned {
		return kernel.New(constant.ErrCodeUserNotActive)
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
		return kernel.New(constant.ErrCodeUserLockedOut)
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
		return kernel.New(constant.ErrCodeUserAlreadyDeleted)
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
		return kernel.New(constant.ErrCodeCredentialNotLocal)
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
