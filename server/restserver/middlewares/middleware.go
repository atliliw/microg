package middlewares

import (
	"github.com/gin-gonic/gin"
)

type ErrorResponseFunc func(c *gin.Context, httpStatus int, err error)

func DefaultErrorResponse(c *gin.Context, httpStatus int, err error) {
	c.JSON(httpStatus, gin.H{
		"code":    httpStatus,
		"message": err.Error(),
	})
}

var errorResponse ErrorResponseFunc = DefaultErrorResponse

func SetErrorResponse(fn ErrorResponseFunc) {
	errorResponse = fn
}

func WriteError(c *gin.Context, httpStatus int, err error) {
	errorResponse(c, httpStatus, err)
}

var Middlewares = defaultMiddlewares()

func defaultMiddlewares() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"recovery": gin.Recovery(),
		"cors":     Cors(),
		"context":  Context(),
	}
}
