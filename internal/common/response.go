package common

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RespondError(c *gin.Context, status int, code string, params ...map[string]any) {
	payload := gin.H{
		"code": code,
	}
	if len(params) > 0 && len(params[0]) > 0 {
		payload["params"] = params[0]
	}
	c.JSON(status, payload)
}

func RespondFromError(c *gin.Context, err error) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		RespondError(c, apiErr.HTTPStatus, apiErr.Code, apiErr.Params)
		return
	}
	RespondError(c, http.StatusInternalServerError, ErrInternalServerError)
}

func RespondSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}
