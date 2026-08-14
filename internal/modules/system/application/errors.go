package application

import "sipon-be/internal/shared/kernel"

// HTTP-semantic codes — dipakai usecase untuk membungkus domain error menjadi
// app error. HTTP layer (httperror) hanya perlu mapping ERR_* → HTTP status.
const (
	ErrCodeBadRequest          kernel.Code = "ERR_BAD_REQUEST"
	ErrCodeUnauthorized        kernel.Code = "ERR_UNAUTHORIZED"
	ErrCodeForbidden           kernel.Code = "ERR_FORBIDDEN"
	ErrCodeNotFound            kernel.Code = "ERR_NOT_FOUND"
	ErrCodeConflict            kernel.Code = "ERR_CONFLICT"
	ErrCodeUnprocessableEntity kernel.Code = "ERR_UNPROCESSABLE_ENTITY"
	ErrCodeInternal            kernel.Code = "ERR_INTERNAL"
)
