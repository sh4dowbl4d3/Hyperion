package race

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"hyperion/backend/internal/httpx"
	"hyperion/backend/internal/labs"
	"hyperion/backend/internal/middleware"
)

type Redeemer interface {
	RedeemVulnerable(ctx context.Context, userID, code string) error
	RedeemSafe(ctx context.Context, userID, code string) error
	StatusForUser(ctx context.Context, userID, code string) (*Status, error)
}

type Lab struct {
	store     Redeemer
	completer Completer
}

func NewLab(store Redeemer, completer Completer) *Lab {
	return &Lab{store: store, completer: completer}
}

func (l *Lab) Meta() labs.Meta {
	return labs.Meta{
		Slug:        LabSlug,
		Name:        "Limited Coupon Drop",
		Description: "A launch coupon can be redeemed exactly five times in total. The redemption endpoint checks remaining capacity and then records the redemption as two separate steps — a window you can race.",
		Objective:   "Send many concurrent redemptions of " + CouponCode + " to /targets/race-condition/redeem and push the global counter past its cap. Confirm the transactional twin at /redeem-safe refuses over-redemption.",
		Category:    "business-logic",
		Difficulty:  labs.DifficultyHard,
		XP:          200,
		VulnerabilityType: "race condition (TOCTOU check-then-act on redemption limit)",
		Hints: []string{
			"The capacity check and the write are not atomic — win the gap between them.",
			"Fire 10+ parallel requests; a single sequential request cannot win the race.",
			"The safe twin reserves capacity with one conditional UPDATE under row lock.",
		},
	}
}

func (l *Lab) RegisterRoutes(group *gin.RouterGroup) error {
	group.GET("/coupon", l.handleStatus)
	group.POST("/redeem", l.handleRedeem(false))
	group.POST("/redeem-safe", l.handleRedeem(true))
	return nil
}

func (l *Lab) handleStatus(c *gin.Context) {
	userID, ok := middleware.UserIDFrom(c)
	if !ok {
		httpx.Unauthorized(c, "no authenticated identity on request")
		return
	}
	st, err := l.store.StatusForUser(c.Request.Context(), userID, CouponCode)
	if err != nil {
		httpx.NotFound(c, "no such coupon")
		return
	}
	c.JSON(http.StatusOK, gin.H{"coupon": st})
}

func (l *Lab) handleRedeem(secure bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := middleware.UserIDFrom(c)
		if !ok {
			httpx.Unauthorized(c, "no authenticated identity on request")
			return
		}
		var req struct {
			Code string `json:"code"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Code == "" {
			httpx.Error(c, http.StatusBadRequest, "invalid_code", "code is required")
			return
		}

		redeem := l.store.RedeemVulnerable
		if secure {
			redeem = l.store.RedeemSafe
		}

		err := redeem(c.Request.Context(), userID, req.Code)
		switch {
		case err == nil:
			c.JSON(http.StatusOK, gin.H{"redeemed": true})
		case err == ErrExhausted:
			httpx.Error(c, http.StatusConflict, "coupon_exhausted", "redemption limit reached")
		case err == ErrAlreadyUsed:
			httpx.Error(c, http.StatusConflict, "already_redeemed", "you already redeemed this coupon")
		case err == ErrCouponNotFound:
			httpx.NotFound(c, "no such coupon")
		default:
			httpx.Internal(c, "unable to redeem coupon")
		}
		l.maybeComplete(c)
	}
}

// Completion: the lab is solved when the vulnerable counter has been pushed
// past its cap (or at least fully consumed while other users were locked out).
func (l *Lab) maybeComplete(c *gin.Context) {
	if l.completer == nil {
		return
	}
	userID, ok := middleware.UserIDFrom(c)
	if !ok {
		return
	}
	st, err := l.store.StatusForUser(c.Request.Context(), userID, CouponCode)
	if err != nil || st == nil || st.Mine != true || OverRedeemed(st) != true {
		return
	}
	if _, err := l.completer.CompleteLab(c.Request.Context(), userID, LabSlug); err != nil {
		c.Error(err)
	}
}
