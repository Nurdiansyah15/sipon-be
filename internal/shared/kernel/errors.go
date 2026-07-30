package kernel

import "fmt"

type Code string

type AppError struct {
	Code Code
	Err  error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code: %s, err: %v", e.Code, e.Err)
	}
	return fmt.Sprintf("code: %s", e.Code)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code Code) *AppError {
	return &AppError{Code: code}
}

func Wrap(code Code, err error) *AppError {
	return &AppError{Code: code, Err: err}
}
