package errors

import "errors"

// RetryableError menandakan kegagalan sementara yang layak di-retry dengan
// backoff dan batas attempt.
type RetryableError struct {
	Err error
}

func NewRetryableError(err error) error {
	return &RetryableError{Err: err}
}

func (e *RetryableError) Error() string {
	if e.Err == nil {
		return "messaging: retryable error"
	}
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error { return e.Err }

// FatalError menandakan kegagalan permanen (payload invalid, routing key tidak
// dikenal, precondition tidak bisa dipenuhi) yang tidak akan membaik dengan retry.
type FatalError struct {
	Err error
}

func NewFatalError(err error) error {
	return &FatalError{Err: err}
}

func (e *FatalError) Error() string {
	if e.Err == nil {
		return "messaging: fatal error"
	}
	return e.Err.Error()
}

func (e *FatalError) Unwrap() error { return e.Err }

func IsFatal(err error) bool {
	var fe *FatalError
	return errors.As(err, &fe)
}

func IsRetryable(err error) bool {
	var re *RetryableError
	return errors.As(err, &re)
}
