package httpx

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

func Error(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, ErrorEnvelope{Error: ErrorBody{Code: code, Message: message}})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "bad_request", message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "unauthorized", message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "forbidden", message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, "not_found", message)
}

func Internal(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, "internal_error", message)
}

var ErrUnauthorized = errors.New("unauthorized")
