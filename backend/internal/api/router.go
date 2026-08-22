package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"moderndvwa/backend/internal/httpx"
	"moderndvwa/backend/internal/middleware"
)

const (
	ServiceName    = "moderndvwa-api"
	ServiceVersion = "0.1.0"

	readinessTimeout = 2 * time.Second
)

type DatabaseChecker interface {
	Ping(ctx context.Context) error
}

func NewRouter(log *slog.Logger, db DatabaseChecker) *gin.Engine {
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
	registerReadiness(v1, db)

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

func registerReadiness(rg *gin.RouterGroup, db DatabaseChecker) {
	rg.GET("/readyz", func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unavailable",
				"checks": gin.H{"database": "not_configured"},
			})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), readinessTimeout)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unavailable",
				"checks": gin.H{"database": "error"},
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"checks": gin.H{"database": "ok"},
		})
	})
}
