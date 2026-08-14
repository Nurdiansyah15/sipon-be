package entity

import (
	"time"

	"sipon-be/internal/modules/identity/domain/user/constant"
	"sipon-be/internal/modules/identity/domain/user/valueobject"
)

type Credential struct {
	ID              string
	UserID          string
	Type            constant.CredentialType
	LoginIdentities []*LoginIdentity
	SecretHash      *valueobject.HashedPassword
	LastChangedAt   *time.Time
	IsPrimary       bool
	UpdatedAt       time.Time
	LastLoginAt     *time.Time
	DeletedAt       *time.Time
}

func NewLocalCredential(id, userID string, hashed valueobject.HashedPassword, isPrimary bool) *Credential {
	now := time.Now()
	return &Credential{
		ID:         id,
		UserID:     userID,
		Type:       constant.CredentialTypeLocal,
		SecretHash: &hashed,
		IsPrimary:  isPrimary,
		UpdatedAt:  now,
	}
}

func NewLocalCredentialWithoutPassword(id, userID string, isPrimary bool) *Credential {
	now := time.Now()
	return &Credential{
		ID:        id,
		UserID:    userID,
		Type:      constant.CredentialTypeLocal,
		IsPrimary: isPrimary,
		UpdatedAt: now,
	}
}

func NewGoogleCredential(id, userID string, isPrimary bool) *Credential {
	now := time.Now()
	return &Credential{
		ID:        id,
		UserID:    userID,
		Type:      constant.CredentialTypeGoogle,
		IsPrimary: isPrimary,
		UpdatedAt: now,
	}
}

func (c *Credential) AddLoginIdentity(identity *LoginIdentity) {
	c.LoginIdentities = append(c.LoginIdentities, identity)
	c.UpdatedAt = time.Now()
}

func (c *Credential) FindLoginIdentity(kind constant.LoginIdentifierKind, value string) *LoginIdentity {
	for _, li := range c.LoginIdentities {
		if li.Kind == kind && li.Value == value && li.DeletedAt == nil {
			return li
		}
	}
	return nil
}

func (c *Credential) FindLoginIdentityByKind(kind constant.LoginIdentifierKind) *LoginIdentity {
	for _, li := range c.LoginIdentities {
		if li.Kind == kind && li.DeletedAt == nil {
			return li
		}
	}
	return nil
}
