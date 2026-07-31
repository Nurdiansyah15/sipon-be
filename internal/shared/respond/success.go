package respond

import "github.com/gin-gonic/gin"

type SuccessBody struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Data       any    `json:"data"`
	Meta       any    `json:"meta"`
}

func Success(c *gin.Context, code int, message string, data any) {
	SuccessWithMeta(c, code, message, data, nil)
}

func SuccessWithMeta(c *gin.Context, code int, message string, data any, meta any) {
	c.JSON(code, SuccessBody{
		Status:     "success",
		StatusCode: code,
		Message:    message,
		Data:       data,
		Meta:       meta,
	})
}

func OK(c *gin.Context, message string, data any) {
	Success(c, 200, message, data)
}

func Created(c *gin.Context, message string, data any) {
	Success(c, 201, message, data)
}
