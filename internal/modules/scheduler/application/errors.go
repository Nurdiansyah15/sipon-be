package application

import "errors"

// RetryableError menandakan kegagalan sementara yang layak di-retry.
type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	if e.Err == nil {
		return "scheduler: retryable error"
	}
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error { return e.Err }

// FatalError menandakan kegagalan permanen (FAILED tanpa retry).
type FatalError struct {
	Err error
}

func (e *FatalError) Error() string {
	if e.Err == nil {
		return "scheduler: fatal error"
	}
	return e.Err.Error()
}

func (e *FatalError) Unwrap() error { return e.Err }

func IsFatal(err error) bool {
	var fe *FatalError
	return errors.As(err, &fe)
}
