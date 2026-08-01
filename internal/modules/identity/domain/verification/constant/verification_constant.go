package constant

import "sipon-be/internal/shared/kernel"

type CodePurpose string

const (
	PurposeEmailVerification CodePurpose = "EMAIL_VERIFICATION"
	PurposePhoneVerification CodePurpose = "PHONE_VERIFICATION"
	PurposeResetPassword     CodePurpose = "RESET_PASSWORD"
	PurposeChangeEmail       CodePurpose = "CHANGE_EMAIL"
	PurposeChangePhone       CodePurpose = "CHANGE_PHONE"
)

const (
	ErrCodeVerificationCodeNotFound     kernel.Code = "VERIFICATION_CODE_NOT_FOUND"
	ErrCodeVerificationCodeExpired      kernel.Code = "VERIFICATION_CODE_EXPIRED"
	ErrCodeVerificationCodeUsed         kernel.Code = "VERIFICATION_CODE_USED"
	ErrCodeVerificationCodeMismatch     kernel.Code = "VERIFICATION_CODE_MISMATCH"
	ErrCodeVerificationInvalidPurpose   kernel.Code = "VERIFICATION_INVALID_PURPOSE"
	ErrCodeTooManyVerificationCode      kernel.Code = "TOO_MANY_VERIFICATION_CODE"
	ErrCodeVerificationNewIdentityEmpty kernel.Code = "VERIFICATION_NEW_IDENTITY_EMPTY"
)
