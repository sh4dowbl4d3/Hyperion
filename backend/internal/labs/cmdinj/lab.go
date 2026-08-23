package cmdinj

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"moderndvwa/backend/internal/httpx"
	"moderndvwa/backend/internal/labs"
	"moderndvwa/backend/internal/middleware"
)

type LookupStore interface {
	LookupVulnerable(ctx context.Context, host string) (*LookupResult, error)
	LookupSafe(ctx context.Context, host string) (*LookupResult, error)
}

type Lab struct {
	store     LookupStore
	completer Completer
}

func NewLab(store LookupStore, completer Completer) *Lab {
	return &Lab{store: store, completer: completer}
}

func (l *Lab) Meta() labs.Meta {
	return labs.Meta{
		Slug:        LabSlug,
		Name:        "Network Tool — DNS Probe",
		Description: "A diagnostics page lets staff run a DNS lookup against a host they type in. The lookup command is assembled by pasting the host straight onto the end of it, and the shell happily runs whatever else it finds.",
		Objective:   "Inject a second command into /targets/command-injection/lookup?host= so the tool executes something beyond the DNS query (try appending ; cat /etc/passwd-style payloads) and recover the leaked output. The hardened twin rejects any host that is not a plain hostname.",
		Category:    "injection",
		Difficulty:  labs.DifficultyMedium,
		XP:          150,
		VulnerabilityType: "command injection (unsanitised input concatenated into a shell command)",
		Hints: []string{
			"The host string is glued onto the end of a command; the shell treats `;` as 'run another'.",
			"Anything after your injected semicolon is executed and its output is appended to the response.",
			"The safe twin allow-lists plain hostnames: letters, digits, dots and dashes only.",
		},
	}
}

func (l *Lab) RegisterRoutes(group *gin.RouterGroup) error {
	group.GET("/lookup", l.handleLookup(l.store.LookupVulnerable))
	group.GET("/lookup-safe", l.handleLookup(l.store.LookupSafe))
	return nil
}

func (l *Lab) handleLookup(
	lookup func(context.Context, string) (*LookupResult, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := middleware.UserIDFrom(c)
		if !ok {
			httpx.Unauthorized(c, "no authenticated identity on request")
			return
		}
		host := strings.TrimSpace(c.Query("host"))
		if host == "" {
			httpx.Error(c, http.StatusBadRequest, "missing_host", "host query parameter is required")
			return
		}
		if len(host) > 300 {
			httpx.Error(c, http.StatusBadRequest, "invalid_host", "host is too long")
			return
		}

		result, err := lookup(c.Request.Context(), host)
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidHost):
				httpx.Error(c, http.StatusBadRequest, "invalid_host",
					"only plain hostnames (letters, digits, dots, dashes) are allowed")
			default:
				httpx.Error(c, http.StatusBadRequest, "invalid_host", "host must not be empty")
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"result": result})

		if l.completer != nil && result.Flagged {
			if _, err := l.completer.CompleteLab(c.Request.Context(), userID, LabSlug); err != nil {
				c.Error(err)
			}
		}
	}
}
