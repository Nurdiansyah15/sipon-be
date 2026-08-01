package entity

import (
	"time"

	"sipon-be/internal/modules/identity/domain/user/valueobject"
	"sipon-be/internal/modules/identity/domain/verification/constant"
	"sipon-be/internal/shared/kernel"
)

type VerificationCode struct {
	ID               string
	UserID           string
	Code             valueobject.OTPCode
	Purpose          constant.CodePurpose
	ExpiresAt        time.Time
	UsedAt           *time.Time
	CreatedAt        time.Time
	NewIdentityValue *string
}

func NewVerificationCode(id, userID, rawCode string, purpose constant.CodePurpose, ttl time.Duration) (*VerificationCode, error) {
	if purpose != constant.PurposeEmailVerification &&
		purpose != constant.PurposePhoneVerification &&
		purpose != constant.PurposeResetPassword &&
		purpose != constant.PurposeChangeEmail &&
		purpose != constant.PurposeChangePhone {
		return nil, kernel.New(constant.ErrCodeVerificationInvalidPurpose)
	}

	code, err := valueobject.NewOTPCode(rawCode)
	if err != nil {
		return nil, err
	}

	return &VerificationCode{
		ID:        id,
		UserID:    userID,
		Code:      code,
		Purpose:   purpose,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
	}, nil
}

func (vc *VerificationCode) Verify(inputCode string) error {
	if vc.UsedAt != nil {
		return kernel.New(constant.ErrCodeVerificationCodeUsed)
	}
	if vc.IsExpired() {
		return kernel.New(constant.ErrCodeVerificationCodeExpired)
	}
	if !vc.Code.Match(inputCode) {
		return kernel.New(constant.ErrCodeVerificationCodeMismatch)
	}
	now := time.Now()
	vc.UsedAt = &now
	return nil
}

func (vc *VerificationCode) IsExpired() bool {
	return time.Now().After(vc.ExpiresAt)
}

func (vc *VerificationCode) SetNewIdentityValue(value string) error {
	if value == "" {
		return kernel.New(constant.ErrCodeVerificationNewIdentityEmpty)
	}
	vc.NewIdentityValue = &value
	return nil
}
