package middleware

// HTTP error code constants used by middleware.
const (
	CodeMissingToken       = "MISSING_TOKEN"
	CodeInvalidTokenFormat = "INVALID_TOKEN_FORMAT"
	CodeInvalidToken       = "INVALID_TOKEN"
	CodeSessionRevoked     = "SESSION_REVOKED"
	CodeForbidden          = "FORBIDDEN"
	CodeInsufficientRole   = "INSUFFICIENT_ROLE"
	CodeInsufficientPerm   = "INSUFFICIENT_PERMISSION"
	CodeRateLimited        = "RATE_LIMITED"
)
