package httperror

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

func ParseValidationErrors(err validator.ValidationErrors) map[string]string {
	errorMessages := make(map[string]string)
	for _, e := range err {
		field := convertToSnakeCase(e.Field())
		errorMessages[field] = validationTagMessage(e.Tag(), e.Param())
	}
	return errorMessages
}

func validationTagMessage(tag, param string) string {
	if param != "" {
		return fmt.Sprintf("%s=%s", tag, param)
	}
	return tag
}

func convertToSnakeCase(str string) string {
	reg := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := reg.ReplaceAllString(str, "${1}_${2}")
	return strings.ToLower(snake)
}
