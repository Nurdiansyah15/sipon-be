package kernel

import "fmt"

type Code string

type AppError struct {
	Code    Code
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code: %s, message: %s, err: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("code: %s, message: %s", e.Code, e.Message)
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

func WrapMsg(code Code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}
