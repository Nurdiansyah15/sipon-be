package domain

import (
	"sipon-be/internal/shared/kernel"
	"time"
)

type CodePurpose string

const (
	PurposeEmailVerification CodePurpose = "EMAIL_VERIFICATION"
	PurposePhoneVerification CodePurpose = "PHONE_VERIFICATION"
	PurposeResetPassword     CodePurpose = "RESET_PASSWORD"
	PurposeChangeEmail       CodePurpose = "CHANGE_EMAIL"
	PurposeChangePhone       CodePurpose = "CHANGE_PHONE"
)

const (
	ErrCodeVerificationCodeNotFound   kernel.Code = "VERIFICATION_CODE_NOT_FOUND"
	ErrCodeVerificationCodeExpired    kernel.Code = "VERIFICATION_CODE_EXPIRED"
	ErrCodeVerificationCodeUsed       kernel.Code = "VERIFICATION_CODE_USED"
	ErrCodeVerificationCodeMismatch   kernel.Code = "VERIFICATION_CODE_MISMATCH"
	ErrCodeVerificationInvalidPurpose kernel.Code = "VERIFICATION_INVALID_PURPOSE"
	ErrCodeTooManyVerificationCode    kernel.Code = "TOO_MANY_VERIFICATION_CODE"
	ErrCodeVerificationNewIdentityEmpty kernel.Code = "VERIFICATION_NEW_IDENTITY_EMPTY"
)

type VerificationCode struct {
	ID               string
	UserID           string
	Code             OTPCode
	Purpose          CodePurpose
	ExpiresAt        time.Time
	UsedAt           *time.Time
	CreatedAt        time.Time
	NewIdentityValue *string
}

func NewVerificationCode(id, userID string, code OTPCode, purpose CodePurpose, expiresAt time.Time) (*VerificationCode, error) {
	if purpose != PurposeEmailVerification &&
		purpose != PurposePhoneVerification &&
		purpose != PurposeResetPassword &&
		purpose != PurposeChangeEmail &&
		purpose != PurposeChangePhone {
		return nil, kernel.New(ErrCodeVerificationInvalidPurpose)
	}

	return &VerificationCode{
		ID:        id,
		UserID:    userID,
		Code:      code,
		Purpose:   purpose,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}, nil
}

func (vc *VerificationCode) Verify(inputCode OTPCode) error {
	if vc.UsedAt != nil {
		return kernel.New(ErrCodeVerificationCodeUsed)
	}
	if vc.IsExpired() {
		return kernel.New(ErrCodeVerificationCodeExpired)
	}
	if vc.Code.String() != inputCode.String() {
		return kernel.New(ErrCodeVerificationCodeMismatch)
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
		return kernel.New(ErrCodeVerificationNewIdentityEmpty)
	}
	vc.NewIdentityValue = &value
	return nil
}
