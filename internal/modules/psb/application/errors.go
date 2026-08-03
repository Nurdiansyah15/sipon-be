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
