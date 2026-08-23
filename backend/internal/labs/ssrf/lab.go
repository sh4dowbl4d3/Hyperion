package ssrf

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"hyperion/backend/internal/httpx"
	"hyperion/backend/internal/labs"
	"hyperion/backend/internal/middleware"
)

const LabSlug = "ssrf"

const (
	fetchTimeout = 4 * time.Second
	maxBodyBytes = 64 * 1024
)

type Completer interface {
	CompleteLab(ctx context.Context, userID, labSlug string) (*labs.CompletionResult, error)
}

// Fetcher abstracts outbound HTTP so tests never touch a real socket.
type Fetcher interface {
	Fetch(ctx context.Context, rawURL string) (int, string, error)
}

// HTTPFetcher performs a plain GET with no scheme/host restrictions — this is
// the vulnerability: whatever URL the client supplies gets fetched.
type HTTPFetcher struct {
	client *http.Client
}

func NewHTTPFetcher() *HTTPFetcher {
	return &HTTPFetcher{client: &http.Client{Timeout: fetchTimeout}}
}

func (f *HTTPFetcher) Fetch(ctx context.Context, rawURL string) (int, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, "", err
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return 0, "", err
	}
	return resp.StatusCode, string(body), nil
}

type Lab struct {
	fetcher   Fetcher
	completer Completer
	baseURL   string // internal synthetic service
}

func NewLab(fetcher Fetcher, completer Completer, internalBaseURL string) *Lab {
	return &Lab{fetcher: fetcher, completer: completer, baseURL: internalBaseURL}
}

func (l *Lab) Meta() labs.Meta {
	return labs.Meta{
		Slug:        LabSlug,
		Name:        "Inventory Webhook",
		Description: "The inventory service offers a URL-preview feature that fetches whatever address it is given. An internal stock-sync service lives on loopback and trusts private callers.",
		Objective:   "Use /targets/ssrf/preview?url= to make the server fetch its own internal service (hint: the webhook config at /targets/ssrf/config reveals where it lives) and read the file under /secret/. The hardened twin validates every URL before fetching.",
		Category:    "request-forgery",
		Difficulty:  labs.DifficultyMedium,
		XP:          150,
		VulnerabilityType: "ssrf (unvalidated server-side URL fetch)",
		Hints: []string{
			"The preview endpoint accepts any absolute URL — including loopback addresses.",
			"Check /config for the internal base URL; the interesting path starts with /secret/.",
			"The safe twin only allows http(s) and rejects loopback, link-local and private ranges.",
		},
	}
}

func (l *Lab) RegisterRoutes(group *gin.RouterGroup) error {
	group.GET("/config", l.handleConfig)
	group.GET("/preview", l.handlePreview(false))
	group.GET("/preview-safe", l.handlePreview(true))
	return nil
}

func (l *Lab) handleConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"webhook": gin.H{
			"name":      "inventory-sync",
			"base_url":  l.baseURL,
			"endpoints": []string{MetadataPath},
			"note":      "internal use only",
		},
	})
}

var ErrBlockedByPolicy = errors.New("url rejected by fetch policy")

func (l *Lab) handlePreview(secure bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := middleware.UserIDFrom(c)
		if !ok {
			httpx.Unauthorized(c, "no authenticated identity on request")
			return
		}

		raw := strings.TrimSpace(c.Query("url"))
		if raw == "" {
			httpx.Error(c, http.StatusBadRequest, "missing_url", "url query parameter is required")
			return
		}
		if len(raw) > 2048 {
			httpx.Error(c, http.StatusBadRequest, "invalid_url", "url too long")
			return
		}

		if secure {
			if err := ValidatePublicURL(raw); err != nil {
				httpx.Error(c, http.StatusBadRequest, "url_blocked", "only public http(s) URLs are allowed")
				return
			}
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), fetchTimeout)
		defer cancel()
		status, body, err := l.fetcher.Fetch(ctx, raw)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error": gin.H{"code": "fetch_failed", "message": "the requested resource could not be fetched"},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": status, "body": body})

		if !secure && l.completer != nil && IsInternalFlag(body) {
			if _, err := l.completer.CompleteLab(c.Request.Context(), userID, LabSlug); err != nil {
				c.Error(err)
			}
		}
	}
}

// ValidatePublicURL is the reference mitigation: scheme allow-list plus
// rejection of loopback, link-local, private and unspecified addresses.
func ValidatePublicURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return ErrBlockedByPolicy
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrBlockedByPolicy
	}
	host := u.Hostname()
	switch strings.ToLower(host) {
	case "", "localhost":
		return ErrBlockedByPolicy
	}
	if ip := netParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
			ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return ErrBlockedByPolicy
		}
		return nil
	}
	// Hostnames containing obvious internal markers are refused as well.
	if strings.HasSuffix(host, ".internal") || strings.HasSuffix(host, ".local") ||
		strings.EqualFold(host, "metadata.google.internal") {
		return ErrBlockedByPolicy
	}
	return nil
}
