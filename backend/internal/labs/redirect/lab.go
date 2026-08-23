package redirect

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"hyperion/backend/internal/httpx"
	"hyperion/backend/internal/labs"
	"hyperion/backend/internal/middleware"
)

const LabSlug = "open-redirect"

// FlagTargetPath is the path on the "internal admin host" whose visit proves
// an open redirect chain was abused. The lab treats any redirect target with
// this path as the flag.
const FlagTargetPath = "/admin/secret-token"

type Completer interface {
	CompleteLab(ctx context.Context, userID, labSlug string) (*labs.CompletionResult, error)
}

type Lab struct {
	completer Completer
}

func NewLab(completer Completer) *Lab {
	return &Lab{completer: completer}
}

func (l *Lab) Meta() labs.Meta {
	return labs.Meta{
		Slug:        LabSlug,
		Name:        "Campaign Link Forwarder",
		Description: "A marketing tracker redirects users through /redirect to count campaign clicks before sending them on their way. The destination arrives as a plain query parameter and is forwarded without question.",
		Objective:   "Abuse the forwarder at /targets/open-redirect/redirect?to= to bounce a victim toward the internal admin console (any host whose path is /admin/secret-token) and read what it exposes. Then verify the hardened twin only lets internal campaign slugs through.",
		Category:    "request-forgery",
		Difficulty:  labs.DifficultyEasy,
		XP:          100,
		VulnerabilityType: "open redirect (unvalidated destination parameter)",
		Hints: []string{
			"The `to` parameter accepts any absolute URL — including hosts that are not campaigns.",
			"You will not actually browse anywhere; the lab inspects where you asked it to send you.",
			"The safe twin only allows relative paths like /promo or /newsletter.",
		},
	}
}

func (l *Lab) RegisterRoutes(group *gin.RouterGroup) error {
	group.GET("/campaigns", l.handleCampaigns)
	group.GET("/redirect", l.handleRedirect(false))
	group.GET("/redirect-safe", l.handleRedirect(true))
	return nil
}

func (l *Lab) handleCampaigns(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"campaigns": []gin.H{
			{"slug": "promo", "name": "Summer promo", "landing": "/promo"},
			{"slug": "newsletter", "name": "Newsletter signup", "landing": "/newsletter"},
		},
	})
}

func (l *Lab) handleRedirect(secure bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := middleware.UserIDFrom(c)
		if !ok {
			httpx.Unauthorized(c, "no authenticated identity on request")
			return
		}
		to := strings.TrimSpace(c.Query("to"))
		if to == "" {
			httpx.Error(c, http.StatusBadRequest, "missing_to", "to query parameter is required")
			return
		}
		if len(to) > 2048 {
			httpx.Error(c, http.StatusBadRequest, "invalid_to", "to is too long")
			return
		}

		if secure && !isInternalSlug(to) {
			httpx.Error(c, http.StatusBadRequest, "to_blocked",
				"only internal campaign paths may be used")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"redirecting_to": to,
			"note":           "In a browser this would issue an HTTP 302 to the target.",
		})

		if !secure && l.completer != nil && targetsFlag(to) {
			if _, err := l.completer.CompleteLab(c.Request.Context(), userID, LabSlug); err != nil {
				c.Error(err)
			}
		}
	}
}

// isInternalSlug allows only root-relative campaign paths in the safe twin.
func isInternalSlug(to string) bool {
	u, err := url.Parse(to)
	if err != nil {
		return false
	}
	if u.Scheme != "" || u.Host != "" || u.Opaque != "" || strings.HasPrefix(to, "//") {
		return false
	}
	switch u.Path {
	case "/promo", "/newsletter":
		return u.RawQuery == "" && u.Fragment == ""
	}
	return false
}

func targetsFlag(to string) bool {
	u, err := url.Parse(to)
	if err != nil {
		return false
	}
	return u.Path == FlagTargetPath
}
