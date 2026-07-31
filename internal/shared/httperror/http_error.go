package httperror

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"sipon-be/internal/shared/kernel"
	"sipon-be/internal/shared/respond"
)

type httpError struct {
	statusCode int
	errorCode  string
	message    string
	details    any
}

func (e *httpError) Error() string {
	return fmt.Sprintf("status: %d, code: %s, message: %s", e.statusCode, e.errorCode, e.message)
}

func badRequest(message string) *httpError {
	return &httpError{statusCode: http.StatusBadRequest, errorCode: "ERR_BAD_REQUEST", message: message}
}

func internalError(message string) *httpError {
	return &httpError{statusCode: http.StatusInternalServerError, errorCode: "ERR_INTERNAL", message: message}
}

func Handle(c *gin.Context, err error) {
	if err == nil {
		return
	}

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		c.Set("request_err", err)
		c.Set("request_err_code", "ERR_UNPROCESSABLE_ENTITY")
		c.Set("request_err_message", "ERR_UNPROCESSABLE_ENTITY")
		payload := ParseValidationErrors(validationErrors)
		respond.Error(c, http.StatusUnprocessableEntity, "ERR_UNPROCESSABLE_ENTITY", payload)
		c.Abort()
		return
	}

	httpErr := mapError(err)
	c.Set("request_err", err)
	c.Set("request_err_code", httpErr.errorCode)
	c.Set("request_err_message", httpErr.message)

	if httpErr.statusCode >= 500 {
		c.Set("internal_err", err)
	}

	payload := any(httpErr.message)
	if httpErr.details != nil {
		payload = httpErr.details
	}
	respond.Error(c, httpErr.statusCode, httpErr.errorCode, payload)
	c.Abort()
}

func mapError(err error) *httpError {
	var kernelError *kernel.AppError
	if errors.As(err, &kernelError) {
		return mapKernelError(kernelError)
	}

	if strings.Contains(err.Error(), "no multipart boundary") {
		return badRequest("invalid or missing multipart boundary in Content-Type header")
	}

	if errors.Is(err, io.EOF) || strings.Contains(err.Error(), "EOF") {
		return badRequest("request body is empty or not provided")
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		return badRequest(fmt.Sprintf("invalid type for field '%s'", typeErr.Field))
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return badRequest("invalid JSON format")
	}

	return internalError("internal server error")
}

func mapKernelError(err *kernel.AppError) *httpError {
	code := string(err.Code)
	msg := err.Message
	if msg == "" {
		msg = string(err.Code)
	}
	statusCode := statusFromCode(code)
	return &httpError{statusCode: statusCode, errorCode: code, message: msg}
}

func statusFromCode(code string) int {
	switch code {
	case "ERR_BAD_REQUEST":
		return http.StatusBadRequest
	case "ERR_UNAUTHORIZED":
		return http.StatusUnauthorized
	case "ERR_FORBIDDEN":
		return http.StatusForbidden
	case "ERR_NOT_FOUND":
		return http.StatusNotFound
	case "ERR_CONFLICT":
		return http.StatusConflict
	case "ERR_GONE":
		return http.StatusGone
	case "ERR_TOO_MANY_REQUESTS":
		return http.StatusTooManyRequests
	case "ERR_UNPROCESSABLE_ENTITY":
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}
