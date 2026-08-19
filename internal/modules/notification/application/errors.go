package application

import (
	"errors"

	"sipon-be/internal/shared/kernel"
)

const (
	ErrCodeBadRequest   kernel.Code = "ERR_BAD_REQUEST"
	ErrCodeNotFound     kernel.Code = "ERR_NOT_FOUND"
	ErrCodeInternal     kernel.Code = "ERR_INTERNAL"
	ErrCodeUnprocessable kernel.Code = "ERR_UNPROCESSABLE_ENTITY"
)

func WrapDomainErr(err error, notFoundCode kernel.Code) error {
	var ke *kernel.AppError
	if errors.As(err, &ke) && ke.Code == notFoundCode {
		return kernel.Wrap(ErrCodeNotFound, err)
	}
	return kernel.Wrap(ErrCodeInternal, err)
}
