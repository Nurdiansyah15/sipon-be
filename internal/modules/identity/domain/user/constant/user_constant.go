package constant

import (
	"time"

	"sipon-be/internal/shared/kernel"
)

type CredentialType string

const (
	CredentialTypeLocal CredentialType = "LOCAL"
)

type LoginIdentifierKind string

const (
	LoginIdentifierKindEmail    LoginIdentifierKind = "EMAIL"
	LoginIdentifierKindPhone    LoginIdentifierKind = "PHONE"
	LoginIdentifierKindUsername LoginIdentifierKind = "USERNAME"
	LoginIdentifierKindNIS      LoginIdentifierKind = "NIS"
)

type LoginIdentityStatus string

const (
	LoginIdentityStatusVerified   LoginIdentityStatus = "VERIFIED"
	LoginIdentityStatusUnverified LoginIdentityStatus = "UNVERIFIED"
)

type UserStatus string

const (
	UserStatusActive UserStatus = "ACTIVE"
	UserStatusBanned UserStatus = "BANNED"
)

const (
	MaxLoginAttempts = 5
	LockoutDuration  = 15 * time.Minute
)

const (
	ErrCodeEmailInvalidFormat         kernel.Code = "EMAIL_INVALID_FORMAT"
	ErrCodeEmailEmpty                 kernel.Code = "EMAIL_EMPTY"
	ErrCodePhoneNumberInvalidFormat   kernel.Code = "PHONE_NUMBER_INVALID_FORMAT"
	ErrCodePhoneNumberEmpty           kernel.Code = "PHONE_NUMBER_EMPTY"
	ErrCodeHashedPasswordTooShort     kernel.Code = "HASHED_PASSWORD_TOO_SHORT"
	ErrCodePlainPasswordTooShort      kernel.Code = "PLAIN_PASSWORD_TOO_SHORT"
	ErrCodePlainPasswordNoUppercase   kernel.Code = "PLAIN_PASSWORD_NO_UPPERCASE"
	ErrCodePlainPasswordNoDigit       kernel.Code = "PLAIN_PASSWORD_NO_DIGIT"
	ErrCodePlainPasswordEmpty         kernel.Code = "PLAIN_PASSWORD_EMPTY"
	ErrCodeOTPCodeInvalidLength       kernel.Code = "OTP_CODE_INVALID_LENGTH"
	ErrCodeOTPCodeNotDigit            kernel.Code = "OTP_CODE_NOT_DIGIT"
	ErrCodeOTPCodeEmpty               kernel.Code = "OTP_CODE_EMPTY"
	ErrCodeUsernameTooShort           kernel.Code = "USERNAME_TOO_SHORT"
	ErrCodeUsernameTooLong            kernel.Code = "USERNAME_TOO_LONG"
	ErrCodeUsernameInvalidChar        kernel.Code = "USERNAME_INVALID_CHAR"
	ErrCodeUsernameEmpty              kernel.Code = "USERNAME_EMPTY"
	ErrCodeLoginIdentifierEmpty       kernel.Code = "LOGIN_IDENTIFIER_EMPTY"
	ErrCodeLoginIdentifierUnknownKind kernel.Code = "LOGIN_IDENTIFIER_UNKNOWN_KIND"
	ErrCodeUserBanned                 kernel.Code = "USER_BANNED"
	ErrCodeUserLockedOut              kernel.Code = "USER_LOCKED_OUT"
	ErrCodeUserNotActive              kernel.Code = "USER_NOT_ACTIVE"
	ErrCodeUserAlreadyActive          kernel.Code = "USER_ALREADY_ACTIVE"
	ErrCodeUserAlreadyBanned          kernel.Code = "USER_ALREADY_BANNED"
	ErrCodeUserAlreadyDeleted         kernel.Code = "USER_ALREADY_DELETED"
	ErrCodeCredentialNotLocal         kernel.Code = "CREDENTIAL_NOT_LOCAL"
	ErrCodeIdentityNotVerified        kernel.Code = "IDENTITY_NOT_VERIFIED"
	ErrCodeNoPrimaryIdentity          kernel.Code = "NO_PRIMARY_IDENTITY"
	ErrCodeUsernameAlreadySet         kernel.Code = "USERNAME_ALREADY_SET"
	ErrCodeInvalidLoginIdentityValue  kernel.Code = "INVALID_LOGIN_IDENTITY_VALUE"
	ErrCodeUserIDRequired             kernel.Code = "USER_ID_REQUIRED"
	ErrCodeUserEmailRequired          kernel.Code = "USER_EMAIL_REQUIRED"
	ErrCodeUserPhoneNumberInvalid     kernel.Code = "USER_PHONE_NUMBER_INVALID"
	ErrCodeUserNotFound               kernel.Code = "USER_NOT_FOUND"
	ErrCodeInternal                   kernel.Code = "USER_INTERNAL_ERROR"
)
