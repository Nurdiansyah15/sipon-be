package entity

import (
	"time"

	"sipon-be/internal/modules/identity/domain/user/constant"
	"sipon-be/internal/modules/identity/domain/user/valueobject"
	"sipon-be/internal/shared/kernel"
)

type LoginIdentity struct {
	ID           string
	UserID       string
	CredentialID string
	Kind         constant.LoginIdentifierKind
	Value        string
	Status       constant.LoginIdentityStatus
	IsPrimary    bool
	VerifiedAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

func NewLoginIdentity(id, userID, credentialID string, kind constant.LoginIdentifierKind, rawValue string, isPrimary bool, verifiedAt *time.Time) (*LoginIdentity, error) {
	normalizedValue, err := valueobject.NormalizeLoginIdentityValue(kind, rawValue)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &LoginIdentity{
		ID:           id,
		UserID:       userID,
		CredentialID: credentialID,
		Kind:         kind,
		Value:        normalizedValue,
		Status:       constant.LoginIdentityStatusUnverified,
		IsPrimary:    isPrimary,
		VerifiedAt:   verifiedAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (li *LoginIdentity) MarkVerified() {
	now := time.Now()
	li.Status = constant.LoginIdentityStatusVerified
	li.VerifiedAt = &now
	li.UpdatedAt = now
}

func (li *LoginIdentity) IsVerified() bool {
	return li.Status == constant.LoginIdentityStatusVerified
}

func (li *LoginIdentity) EnsureVerified() error {
	if !li.IsVerified() {
		return kernel.WrapMsg(constant.ErrCodeIdentityNotVerified, "Identitas belum diverifikasi", nil)
	}
	return nil
}
