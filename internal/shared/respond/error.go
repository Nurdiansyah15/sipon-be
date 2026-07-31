package respond

import "github.com/gin-gonic/gin"

type ErrorBody struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	ErrorCode  string `json:"error_code"`
	Errors     any    `json:"errors"`
}

func Error(c *gin.Context, code int, errorCode string, errors any) {
	c.JSON(code, ErrorBody{
		Status:     "error",
		StatusCode: code,
		ErrorCode:  errorCode,
		Errors:     errors,
	})
}

func AbortWithError(c *gin.Context, code int, errorCode string, errors any) {
	c.AbortWithStatusJSON(code, ErrorBody{
		Status:     "error",
		StatusCode: code,
		ErrorCode:  errorCode,
		Errors:     errors,
	})
}
