package core

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"microg/pkg/errors"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"msg"`
	Detail    string `json:"detail,omitempty"`
	Reference string `json:"reference,omitempty"`
}

func WriteResponse(c *gin.Context, err error, data interface{}) {
	if err == nil {
		c.JSON(http.StatusOK, Response{
			Code:    0,
			Message: "OK",
			Data:    data,
		})
		return
	}

	parseCoder := errors.FromGrpcError(err)
	coder := errors.ParseCoder(parseCoder)
	if coder != nil {
		status := coder.HTTPStatus()
		c.JSON(status, ErrResponse{
			Code:      coder.Code(),
			Message:   coder.String(),
			Detail:    fmt.Sprintf("%#+v", parseCoder),
			Reference: coder.Reference(),
		})
		return
	}

	httpStatus, errCode, message := parseError(err)
	c.JSON(httpStatus, ErrResponse{
		Code:    errCode,
		Message: message,
		Detail:  fmt.Sprintf("%#+v", err),
	})
}

func parseError(err error) (int, int, string) {
	switch err.(type) {
	case validator.ValidationErrors:
		return http.StatusBadRequest, 400, "参数验证失败"
	default:
		return http.StatusInternalServerError, 500, err.Error()
	}
}
