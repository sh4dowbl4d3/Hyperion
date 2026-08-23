package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"moderndvwa/backend/internal/auth"
	"moderndvwa/backend/internal/httpx"
	"moderndvwa/backend/internal/labs"
	"moderndvwa/backend/internal/middleware"
)

type CatalogSource interface {
	List(ctx context.Context) ([]labs.CatalogEntry, error)
	Get(ctx context.Context, slug string) (*labs.CatalogEntry, error)
}

type ProgressSource interface {
	ForUser(ctx context.Context, userID string) ([]labs.Record, error)
	SummaryForUser(ctx context.Context, userID string) (*labs.Summary, error)
}

type LabsHandler struct {
	catalog  CatalogSource
	progress ProgressSource
	log      *slog.Logger
}

func NewLabsHandler(catalog CatalogSource, progress ProgressSource, log *slog.Logger) *LabsHandler {
	return &LabsHandler{catalog: catalog, progress: progress, log: log}
}

type labSummaryDTO struct {
	Slug        string     `json:"slug"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	Difficulty  string     `json:"difficulty"`
	XP          int        `json:"xp"`
	Status      string     `json:"status"`
	XPAwarded   int        `json:"xp_awarded"`
	CompletedAt *time.Time `json:"completed_at"`
}

type labDetailDTO struct {
	labSummaryDTO
	Objective string   `json:"objective"`
	Hints     []string `json:"hints"`

	VulnerabilityType *string `json:"vulnerability_type"`
}

func (h *LabsHandler) List(c *gin.Context) {
	userID, ok := middleware.UserIDFrom(c)
	if !ok {
		httpx.Unauthorized(c, "no authenticated identity on request")
		return
	}

	entries, err := h.catalog.List(c.Request.Context())
	if err != nil {
		h.log.Error("catalog list failed", slog.String("error", err.Error()))
		httpx.Internal(c, "unable to load labs")
		return
	}

	records, err := h.progress.ForUser(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("progress lookup failed", slog.String("error", err.Error()))
		httpx.Internal(c, "unable to load labs")
		return
	}
	bySlug := make(map[string]labs.Record, len(records))
	for _, rec := range records {
		bySlug[rec.LabSlug] = rec
	}

	items := make([]labSummaryDTO, 0, len(entries))
	for _, entry := range entries {
		items = append(items, toSummaryDTO(entry, bySlug[entry.Slug]))
	}
	c.JSON(http.StatusOK, gin.H{"labs": items})
}

func (h *LabsHandler) Detail(c *gin.Context) {
	userID, ok := middleware.UserIDFrom(c)
	if !ok {
		httpx.Unauthorized(c, "no authenticated identity on request")
		return
	}
	slug := c.Param("slug")

	entry, err := h.catalog.Get(c.Request.Context(), slug)
	if err != nil {
		h.log.Error("catalog lookup failed", slog.String("slug", slug), slog.String("error", err.Error()))
		httpx.NotFound(c, "no lab with that id exists")
		return
	}

	var record labs.Record
	records, err := h.progress.ForUser(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("progress lookup failed", slog.String("error", err.Error()))
		httpx.Internal(c, "unable to load lab")
		return
	}
	for _, rec := range records {
		if rec.LabSlug == slug {
			record = rec
			break
		}
	}

	dto := labDetailDTO{
		labSummaryDTO: toSummaryDTO(*entry, record),
		Objective:     entry.Objective,
		Hints:         entry.Hints,
	}
	if record.Status == labs.StatusCompleted {
		vulnType := entry.VulnerabilityType
		dto.VulnerabilityType = &vulnType
	}
	c.JSON(http.StatusOK, gin.H{"lab": dto})
}

func (h *LabsHandler) ProgressOverview(c *gin.Context) {
	userID, ok := middleware.UserIDFrom(c)
	if !ok {
		httpx.Unauthorized(c, "no authenticated identity on request")
		return
	}

	summary, err := h.progress.SummaryForUser(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("progress summary failed", slog.String("error", err.Error()))
		httpx.Internal(c, "unable to load progress")
		return
	}
	records, err := h.progress.ForUser(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("progress lookup failed", slog.String("error", err.Error()))
		httpx.Internal(c, "unable to load progress")
		return
	}
	c.JSON(http.StatusOK, gin.H{"summary": summary, "records": records})
}

func toSummaryDTO(entry labs.CatalogEntry, record labs.Record) labSummaryDTO {
	dto := labSummaryDTO{
		Slug:        entry.Slug,
		Name:        entry.Name,
		Description: entry.Description,
		Category:    entry.Category,
		Difficulty:  string(entry.Difficulty),
		XP:          entry.XP,
		Status:      string(labs.StatusNotStarted),
	}
	if record.LabSlug == entry.Slug && record.Status != "" {
		dto.Status = string(record.Status)
		dto.XPAwarded = record.XPAwarded
		dto.CompletedAt = record.CompletedAt
	}
	return dto
}

var ErrLabNotFound = errors.New("lab not found")

func RegisterLabRoutes(rg *gin.RouterGroup, tokens *auth.TokenService, catalog CatalogSource, progress ProgressSource, log *slog.Logger) {
	handler := NewLabsHandler(catalog, progress, log)

	labRoutes := rg.Group("/labs", middleware.Authenticate(tokens))
	labRoutes.GET("", handler.List)
	labRoutes.GET("/:slug", handler.Detail)

	progressRoutes := rg.Group("/progress", middleware.Authenticate(tokens))
	progressRoutes.GET("", handler.ProgressOverview)
}
