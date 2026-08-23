package jwtlab

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"hyperion/backend/internal/httpx"
	"hyperion/backend/internal/labs"
	"hyperion/backend/internal/middleware"
)

const LabSlug = "jwt"

// AdminAction is what the forged admin token unlocks — the flag payload.
const AdminAction = FlagPayload

type Completer interface {
	CompleteLab(ctx context.Context, userID, labSlug string) (*labs.CompletionResult, error)
}

type Lab struct {
	completer Completer
	now       func() time.Time
}

func NewLab(completer Completer) *Lab {
	return &Lab{completer: completer, now: time.Now}
}

func (l *Lab) Meta() labs.Meta {
	return labs.Meta{
		Slug:        LabSlug,
		Name:        "Guest Badge Forge",
		Description: "A legacy badge printer issues JWTs signed with a well-known weak secret and honours whatever algorithm the token declares. Escalate your guest badge to an admin badge and open the restricted panel.",
		Objective:   "Take the guest token from /targets/jwt/guest-token, craft an admin-privileged variant that /targets/jwt/admin-panel accepts (unsigned alg:none or self-signed with the weak secret), and retrieve the flag. The hardened verifier behind /admin-panel-safe rejects every forgery.",
		Category:    "authentication",
		Difficulty:  labs.DifficultyMedium,
		XP:          150,
		VulnerabilityType: "jwt forgery (alg:none acceptance + weak HMAC secret)",
		Hints: []string{
			"Decode the guest token's header — the signature scheme is declared by the token itself.",
			"The signing secret is 'secret'. Try re-signing your own claims, or dropping the signature entirely.",
			"Keep sub, iss and exp sensible; the panel checks role == 'admin' only.",
		},
	}
}

func (l *Lab) RegisterRoutes(group *gin.RouterGroup) error {
	group.GET("/guest-token", l.handleGuestToken)
	group.GET("/admin-panel", l.handleAdminPanel(false))
	group.GET("/admin-panel-safe", l.handleAdminPanel(true))
	return nil
}

func (l *Lab) handleGuestToken(c *gin.Context) {
	userID, ok := middleware.UserIDFrom(c)
	if !ok {
		httpx.Unauthorized(c, "no authenticated identity on request")
		return
	}
	token, err := IssueVulnerable(labSubject(userID), l.now())
	if err != nil {
		httpx.Internal(c, "unable to mint badge")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"subject": labSubject(userID),
		"note":    "Legacy badge service. Handle with curiosity.",
	})
}

func labSubject(platformUserID string) string {
	// The lab flow uses its own subject namespace so platform identities are
	// never confused with lab tokens.
	return "badge-" + platformUserID
}

func (l *Lab) handleAdminPanel(secure bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := middleware.UserIDFrom(c)
		if !ok {
			httpx.Unauthorized(c, "no authenticated identity on request")
			return
		}
		authHeader := c.GetHeader("X-Lab-Badge")
		raw := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if raw == "" {
			httpx.Error(c, http.StatusBadRequest, "missing_badge", "send your badge in X-Lab-Badge")
			return
		}

		var claims *Claims
		if secure {
			v := &SecureVerifier{Secret: StrongSecret, Issuer: IssuerLab, Now: l.now}
			verified, err := v.Verify(raw)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": gin.H{"code": "invalid_badge", "message": "badge rejected"},
				})
				return
			}
			claims = verified
		} else {
			verified, err := DecodeVulnerable(raw)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": gin.H{"code": "invalid_badge", "message": "badge rejected"},
				})
				return
			}
			claims = verified
		}

		if claims.Role != "admin" {
			httpx.Forbidden(c, "admin badge required for this panel")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"panel":  "restricted operations",
			"granted_to": claims.Sub,
			"result":     AdminAction,
		})

		// Completion: forging succeeded on the vulnerable path only.
		if !secure && l.completer != nil && wasForged(c, raw) {
			if _, err := l.completer.CompleteLab(c.Request.Context(), userID, LabSlug); err != nil {
				c.Error(err)
			}
		}
	}
}

// wasForged reports whether the accepted badge differs from anything the lab
// itself would issue for this user (i.e. the learner crafted it).
func wasForged(c *gin.Context, raw string) bool {
	userID, _ := middleware.UserIDFrom(c)
	issued, err := IssueVulnerable(labSubject(userID), time.Unix(0, 0))
	if err != nil {
		return false
	}
	return raw != issued
}

// StrongSecret backs the secure reference verifier. It is separate from both
// the platform JWT secret and the lab's weak secret.
var StrongSecret = "lab-reference-verifier-secret-change-me"
