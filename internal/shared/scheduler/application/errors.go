package application

import "errors"

var (
	ErrHandlerNotFound = errors.New("scheduler: handler tidak terdaftar untuk job type ini")
)
