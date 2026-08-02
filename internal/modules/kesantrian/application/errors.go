package application

import (
	"errors"

	"sipon-be/internal/shared/kernel"
)

// HTTP-semantic codes — dipakai usecase untuk membungkus domain error menjadi app error.
// HTTP layer (httperror) hanya perlu mapping ERR_* → HTTP status.
const (
	ErrCodeBadRequest          kernel.Code = "ERR_BAD_REQUEST"
	ErrCodeUnauthorized        kernel.Code = "ERR_UNAUTHORIZED"
	ErrCodeForbidden           kernel.Code = "ERR_FORBIDDEN"
	ErrCodeNotFound            kernel.Code = "ERR_NOT_FOUND"
	ErrCodeConflict            kernel.Code = "ERR_CONFLICT"
	ErrCodeUnprocessableEntity kernel.Code = "ERR_UNPROCESSABLE_ENTITY"
	ErrCodeInternal            kernel.Code = "ERR_INTERNAL"
)

// WrapRepoErr maps a domain-level "not found" kernel.Code to
// ErrCodeNotFound, and anything else to ErrCodeInternal. Kept as a small
// shared helper (unlike identity, which repeats this switch inline in every
// use case) purely to avoid re-typing the same errors.As boilerplate across
// kesantrian's ~15 use cases.
func WrapRepoErr(err error, notFoundCode kernel.Code) error {
	var ke *kernel.AppError
	if errors.As(err, &ke) && ke.Code == notFoundCode {
		return kernel.Wrap(ErrCodeNotFound, err)
	}
	return kernel.Wrap(ErrCodeInternal, err)
}
