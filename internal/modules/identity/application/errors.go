package application

import "sipon-be/internal/shared/kernel"

const (
	ErrCodeUserNotFound       kernel.Code = "USER_NOT_FOUND"
	ErrCodeDuplicateEmail     kernel.Code = "DUPLICATE_EMAIL"
	ErrCodeDuplicatePhone     kernel.Code = "DUPLICATE_PHONE"
	ErrCodeDuplicateUsername  kernel.Code = "DUPLICATE_USERNAME"
	ErrCodeInvalidCredentials kernel.Code = "INVALID_CREDENTIALS"
	ErrCodeSessionRevoked     kernel.Code = "SESSION_REVOKED"
	ErrCodeInvalidToken       kernel.Code = "INVALID_TOKEN"
	ErrCodeTokenExpired       kernel.Code = "TOKEN_EXPIRED"
	ErrCodeRateLimited        kernel.Code = "RATE_LIMITED"
)
