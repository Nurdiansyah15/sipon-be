package entity_test

import (
	"errors"

	"sipon-be/internal/shared/kernel"
)

func asAppError(err error, target **kernel.AppError) bool {
	return errors.As(err, target)
}
