package idor

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"hyperion/backend/internal/httpx"
	"hyperion/backend/internal/labs"
	"hyperion/backend/internal/middleware"
)

const LabSlug = "idor"

// FlagContent is stored in another user's confidential document. Completion
// requires reading that document through the vulnerable endpoint.
const FlagContent = "FLAG-IDOR-3f8a92: quarterly access review"

type Completer interface {
	CompleteLab(ctx context.Context, userID, labSlug string) (*labs.CompletionResult, error)
}

// DocumentStore is the minimal store surface the handlers need; tests supply
// fakes. GetByID intentionally has no ownership filter — authorization is the
// handler's job, which is exactly what this lab teaches.
type DocumentStore interface {
	ListForUser(ctx context.Context, userID string) ([]Document, error)
	GetByID(ctx context.Context, id int) (*Document, error)
}

type Lab struct {
	store     DocumentStore
	completer Completer
}

func NewLab(store DocumentStore, completer Completer) *Lab {
	return &Lab{store: store, completer: completer}
}

func (l *Lab) Meta() labs.Meta {
	return labs.Meta{
		Slug:        LabSlug,
		Name:        "Shared Drive Documents",
		Description: "The document service resolves files by sequential numeric ID. One of your colleagues keeps a confidential review on the same drive — the API trusts whoever asks.",
		Objective:   "Enumerate document IDs on /targets/idor/documents/:id and read a document you do not own (classification 'confidential'). Then confirm /targets/idor/documents-safe/:id refuses cross-tenant reads.",
		Category:    "authorization",
		Difficulty:  labs.DifficultyEasy,
		XP:          100,
		VulnerabilityType: "idor/bola (missing object-level authorization)",
		Hints: []string{
			"Your own listing reveals how IDs are shaped — they are small and sequential.",
			"Try IDs just above and below the ones you own.",
			"The secure twin checks owner_id == your user id before returning content.",
		},
	}
}

func (l *Lab) RegisterRoutes(group *gin.RouterGroup) error {
	group.GET("/documents", l.handleList)
	group.GET("/documents/:id", l.handleGet(false))
	group.GET("/documents-safe/:id", l.handleGet(true))
	return nil
}

func (l *Lab) handleList(c *gin.Context) {
	userID, ok := middleware.UserIDFrom(c)
	if !ok {
		httpx.Unauthorized(c, "no authenticated identity on request")
		return
	}
	if seeder, ok := l.store.(interface {
		EnsureSeeded(context.Context, string) error
	}); ok {
		if err := seeder.EnsureSeeded(c.Request.Context(), userID); err != nil {
			httpx.Internal(c, "unable to load documents")
			return
		}
	}
	docs, err := l.store.ListForUser(c.Request.Context(), userID)
	if err != nil {
		httpx.Internal(c, "unable to load documents")
		return
	}
	c.JSON(http.StatusOK, gin.H{"documents": docs})
}

func (l *Lab) handleGet(enforceOwnership bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := middleware.UserIDFrom(c)
		if !ok {
			httpx.Unauthorized(c, "no authenticated identity on request")
			return
		}

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil || id < 1 || id > 1000000 {
			httpx.Error(c, http.StatusBadRequest, "invalid_id", "document id must be a positive integer")
			return
		}

		doc, err := l.store.GetByID(c.Request.Context(), id)
		if err != nil {
			// Unknown vs forbidden collapses to not-found: never confirm existence.
			httpx.NotFound(c, "no such document")
			return
		}

		if enforceOwnership && doc.OwnerID != userID {
			httpx.NotFound(c, "no such document")
			return
		}

		c.JSON(http.StatusOK, gin.H{"document": doc})

		if !enforceOwnership && l.completer != nil && doc.OwnerID != userID &&
			strings.EqualFold(doc.Classification, "confidential") {
			if _, err := l.completer.CompleteLab(c.Request.Context(), userID, LabSlug); err != nil {
				c.Error(err)
			}
		}
	}
}
