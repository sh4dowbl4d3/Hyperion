package api

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"moderndvwa/backend/internal/httpx"
	"moderndvwa/backend/internal/middleware"
)

const (
	ServiceName    = "moderndvwa-api"
	ServiceVersion = "0.1.0"
)

func NewRouter(log *slog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.HandleMethodNotAllowed = true

	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))

	r.NoRoute(func(c *gin.Context) {
		httpx.NotFound(c, "resource not found")
	})
	r.NoMethod(func(c *gin.Context) {
		httpx.Error(c, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	})

	v1 := r.Group("/api/v1")
	registerHealth(v1)

	return r
}

func registerHealth(rg *gin.RouterGroup) {
	rg.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": ServiceName,
			"version": ServiceVersion,
		})
	})
}
