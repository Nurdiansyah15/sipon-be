package application

import (
	"errors"

	"sipon-be/internal/shared/kernel"
)

const (
	ErrCodeBadRequest          kernel.Code = "ERR_BAD_REQUEST"
	ErrCodeUnauthorized        kernel.Code = "ERR_UNAUTHORIZED"
	ErrCodeForbidden           kernel.Code = "ERR_FORBIDDEN"
	ErrCodeNotFound            kernel.Code = "ERR_NOT_FOUND"
	ErrCodeConflict            kernel.Code = "ERR_CONFLICT"
	ErrCodeUnprocessableEntity kernel.Code = "ERR_UNPROCESSABLE_ENTITY"
	ErrCodeInternal            kernel.Code = "ERR_INTERNAL"
)

func WrapRepoErr(err error, notFoundCode kernel.Code) error {
	var ke *kernel.AppError
	if errors.As(err, &ke) && ke.Code == notFoundCode {
		return kernel.Wrap(ErrCodeNotFound, err)
	}
	return kernel.Wrap(ErrCodeInternal, err)
}

func WrapConflictErr(err error, conflictCode kernel.Code) error {
	var ke *kernel.AppError
	if errors.As(err, &ke) && ke.Code == conflictCode {
		return kernel.Wrap(ErrCodeConflict, err)
	}
	return kernel.Wrap(ErrCodeInternal, err)
}

func WrapBadRequestErr(err error, badRequestCode kernel.Code) error {
	var ke *kernel.AppError
	if errors.As(err, &ke) && ke.Code == badRequestCode {
		return kernel.Wrap(ErrCodeUnprocessableEntity, err)
	}
	return kernel.Wrap(ErrCodeInternal, err)
}

// IsNotFoundErr reports whether err is the given domain not-found code
// (e.g. raised by a repository Find* helper when a row is absent).
func IsNotFoundErr(err error, notFoundCode kernel.Code) bool {
	var ke *kernel.AppError
	return errors.As(err, &ke) && ke.Code == notFoundCode
}
