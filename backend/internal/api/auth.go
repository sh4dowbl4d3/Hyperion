package api

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"moderndvwa/backend/internal/auth"
	"moderndvwa/backend/internal/httpx"
	"moderndvwa/backend/internal/middleware"
	"moderndvwa/backend/internal/users"
)

type AuthHandler struct {
	service *auth.Service
	tokens  *auth.TokenService
	store   auth.UserStore
	log     *slog.Logger
}

func NewAuthHandler(service *auth.Service, tokens *auth.TokenService, store auth.UserStore, log *slog.Logger) *AuthHandler {
	return &AuthHandler{service: service, tokens: tokens, store: store, log: log}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func toUserResponse(u *users.User) userResponse {
	return userResponse{ID: u.ID, Email: u.Email, Role: u.Role, CreatedAt: u.CreatedAt}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, "email and password are required")
		return
	}

	user, err := h.service.Register(c.Request.Context(), auth.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
	})
	switch {
	case errors.Is(err, auth.ErrInvalidEmail):
		httpx.BadRequest(c, "provide a valid email address")
	case errors.Is(err, auth.ErrWeakPassword):
		httpx.BadRequest(c, "password must be between 10 and 72 characters")
	case errors.Is(err, auth.ErrEmailTaken):
		httpx.Error(c, http.StatusConflict, "email_taken", "this email is already registered")
	case err != nil:
		h.log.Error("registration failed", slog.String("error", err.Error()))
		httpx.Internal(c, "unable to complete registration")
	default:
		c.JSON(http.StatusCreated, gin.H{"user": toUserResponse(user)})
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, "email and password are required")
		return
	}

	user, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		httpx.Error(c, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
		return
	}
	if err != nil {
		h.log.Error("login failed", slog.String("error", err.Error()))
		httpx.Internal(c, "unable to complete login")
		return
	}

	token, expiresAt, err := h.tokens.Issue(user.ID, user.Role)
	if err != nil {
		h.log.Error("token issuance failed", slog.String("error", err.Error()))
		httpx.Internal(c, "unable to complete login")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token":      token,
		"expires_at": expiresAt,
		"user":       toUserResponse(user),
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	id, ok := middleware.UserIDFrom(c)
	if !ok {
		httpx.Unauthorized(c, "no authenticated identity on request")
		return
	}
	user, err := h.store.ByID(c.Request.Context(), id)
	if errors.Is(err, users.ErrNotFound) {
		httpx.Unauthorized(c, "authenticated account no longer exists")
		return
	}
	if err != nil {
		h.log.Error("current user lookup failed", slog.String("error", err.Error()))
		httpx.Internal(c, "unable to load account")
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": toUserResponse(user)})
}

func RegisterAuthRoutes(rg *gin.RouterGroup, service *auth.Service, tokens *auth.TokenService, store auth.UserStore, log *slog.Logger) {
	handler := NewAuthHandler(service, tokens, store, log)

	authGroup := rg.Group("/auth")
	authGroup.POST("/register", handler.Register)
	authGroup.POST("/login", handler.Login)
	authGroup.GET("/me", middleware.Authenticate(tokens), handler.Me)

	admin := rg.Group("/admin", middleware.Authenticate(tokens), middleware.RequireRole("admin"))
	admin.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "scope": "admin"})
	})
}
