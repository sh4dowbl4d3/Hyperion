package xss

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
		Name:        "Feedback Wall",
		Description: "A product feedback wall stores and re-serves visitor comments verbatim. Plant a script payload in a comment and get it to execute for every reader of the wall.",
		Objective:   "Post a comment containing " + FlagMarker + " through /targets/xss/comments, then load /targets/xss/comments/feed — the vulnerable feed serves it back as raw markup. The sanitized twin at /targets/xss/comments-safe neutralises the payload.",
		Category:    "injection",
		Difficulty:  labs.DifficultyEasy,
		XP:          100,
		VulnerabilityType: "stored-xss (unescaped user content served as HTML)",
		Hints: []string{
			"Comments are stored verbatim and returned verbatim by the feed.",
			"The completion check looks for the exact marker string inside any stored comment body.",
			"Compare the raw body field from /comments with the rendered field from /comments-safe.",
		},
	}
}

func (l *Lab) RegisterRoutes(group *gin.RouterGroup) error {
	group.GET("/comments", l.handleList(l.store.ListVulnerable))
	group.POST("/comments", l.handleCreate(l.store.CreateVulnerable))
	group.GET("/comments/feed", l.handleFeed(l.store.ListVulnerable))

	group.GET("/comments-safe", l.handleList(l.store.ListSafe))
	group.POST("/comments-safe", l.handleCreate(l.store.CreateSafe))
	group.GET("/comments-safe/feed", l.handleFeed(l.store.ListSafe))
	return nil
}

func (l *Lab) handleList(list func(context.Context) ([]Comment, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		comments, err := list(c.Request.Context())
		if err != nil {
			httpx.Internal(c, "unable to load comments")
			return
		}
		c.JSON(http.StatusOK, gin.H{"comments": comments})
	}
}

func (l *Lab) handleCreate(create func(context.Context, string, string) (*Comment, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Author string `json:"author"`
			Body   string `json:"body"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			httpx.Error(c, http.StatusBadRequest, "invalid_body", "request body must be JSON with author and body")
			return
		}
		author := strings.TrimSpace(req.Author)
		body := strings.TrimSpace(req.Body)
		if author == "" || len(author) > 80 {
			httpx.Error(c, http.StatusBadRequest, "invalid_author", "author must be 1-80 characters")
			return
		}
		if body == "" || len(body) > 2000 {
			httpx.Error(c, http.StatusBadRequest, "invalid_comment", "body must be 1-2000 characters")
			return
		}
		if strings.Contains(body, "\x00") {
			httpx.Error(c, http.StatusBadRequest, "invalid_comment", "body contains invalid characters")
			return
		}

		comment, err := create(c.Request.Context(), author, body)
		if err != nil {
			httpx.Internal(c, "unable to store comment")
			return
		}
		c.JSON(http.StatusCreated, gin.H{"comment": comment})
		l.maybeComplete(c)
	}
}

// handleFeed renders the wall as an HTML page. The vulnerable variant writes
// stored bodies straight into markup; this is where stored XSS executes.
func (l *Lab) handleFeed(list func(context.Context) ([]Comment, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		comments, err := list(c.Request.Context())
		if err != nil {
			httpx.Internal(c, "unable to load comments")
			return
		}
		var b strings.Builder
		b.WriteString("<!doctype html><html><head><title>Feedback wall</title></head><body>")
		b.WriteString("<h1>Feedback wall</h1><ul id=\"wall\">")
		for _, cm := range comments {
			b.WriteString("<li data-id=\"")
			b.WriteString(escapeHTML(intToString(cm.ID)))
			b.WriteString("\"><strong>")
			b.WriteString(escapeHTML(cm.Author)) // author is always escaped; only body differs
			b.WriteString(":</strong> ")
			b.WriteString(bodyForFeed(cm)) // VULNERABLE vs SAFE divergence lives here
			b.WriteString("</li>")
		}
		b.WriteString("</ul></body></html>")

		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(b.String()))
		l.maybeComplete(c)
	}
}

func (l *Lab) maybeComplete(c *gin.Context) {
	if l.completer == nil {
		return
	}
	comments, err := l.store.ListVulnerable(c.Request.Context())
	if err != nil || !AnyBodyContains(comments, FlagMarker) {
		return
	}
	userID, ok := middleware.UserIDFrom(c)
	if !ok {
		return
	}
	if _, err := l.completer.CompleteLab(c.Request.Context(), userID, LabSlug); err != nil {
		c.Error(err)
	}
}

func AnyBodyContains(comments []Comment, needle string) bool {
	for _, cm := range comments {
		if strings.Contains(cm.Body, needle) {
			return true
		}
	}
	return false
}

func escapeHTML(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&#34;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits [20]byte
	i := len(digits)
	for n > 0 {
		i--
		digits[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		digits[i] = '-'
	}
	return string(digits[i:])
}
