package sqli

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"moderndvwa/backend/internal/httpx"
	"moderndvwa/backend/internal/labs"
	"moderndvwa/backend/internal/middleware"
)

type Lab struct {
	store     Store
	completer Completer
}

func NewLab(store Store, completer Completer) *Lab {
	return &Lab{store: store, completer: completer}
}

func (l *Lab) Meta() labs.Meta {
	return labs.Meta{
		Slug:        LabSlug,
		Name:        "Contact Directory Search",
		Description: "A staff directory search endpoint builds its SQL query by string concatenation. Extract the hidden archive record that the WHERE clause is supposed to keep out of reach.",
		Objective:   "Use the /targets/sqli/search endpoint to reveal the hidden 'Archive Vault' contact and recover its flag note. Then compare with /targets/sqli/search-safe, which is immune.",
		Category:    "injection",
		Difficulty:  labs.DifficultyEasy,
		XP:          100,
		VulnerabilityType: "sql-injection (string-concatenated ILIKE clause)",
		Hints: []string{
			"The q parameter is pasted directly inside a quoted string in the SQL text.",
			"A single quote closes the literal; everything after it is interpreted as SQL.",
			"' OR true -- is the classic opener when the input lands inside a WHERE clause.",
		},
	}
}

func (l *Lab) RegisterRoutes(group *gin.RouterGroup) error {
	group.GET("/search", l.handleSearch(l.store.SearchVulnerable, true))
	group.GET("/search-safe", l.handleSearch(l.store.SearchSafe, false))
	return nil
}

func (l *Lab) handleSearch(search func(context.Context, string) ([]Contact, error), awardsCompletion bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := strings.TrimSpace(c.Query("q"))
		if q == "" {
			httpx.Error(c, http.StatusBadRequest, "invalid_query", "q must not be empty")
			return
		}
		if len(q) > 200 {
			httpx.Error(c, http.StatusBadRequest, "invalid_query", "q must be at most 200 characters")
			return
		}

		contacts, err := search(c.Request.Context(), q)
		if err != nil {
			// The vulnerable path can produce arbitrary SQL errors from
			// malformed input. Never echo driver details to the client.
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "search_failed",
					"message": "the directory search could not complete",
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"contacts": contacts})

		if awardsCompletion && l.completer != nil && ContainsFlag(contacts) {
			userID, ok := middleware.UserIDFrom(c)
			if ok {
				if _, err := l.completer.CompleteLab(c.Request.Context(), userID, LabSlug); err != nil {
					c.Error(err) // logged by middleware; completion is retried on next success
				}
			}
		}
	}
}
