package otp

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/middleware"
)

// RegisterRoutes adds OTP endpoints under /auth/otp.
// /send gets a dedicated stricter per-IP limiter — the service-level
// 30s cooldown is per-target (phone/email) so a scripted list of 60
// unique numbers/min would sneak past it and hit our SMS gateway.
// At ~5 sends per IP per 5 min a legitimate user can retry after a
// typo, but a bot can't enumerate numbers to burn our gateway budget.
func RegisterRoutes(r *gin.Engine) {
	strict := middleware.RateLimit(middleware.RateLimiterConfig{
		Rate:  1.0 / 60.0, // one token per minute
		Burst: 5,
		KeyFunc: func(c *gin.Context) string { return c.ClientIP() },
	})
	g := r.Group("/auth/otp")
	{
		g.POST("/send", strict, handleSend)
		g.POST("/verify", handleVerify)
	}
}

func handleSend(c *gin.Context) {
	var in SendInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	if in.Phone == "" && in.Email == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "phone or email required"})
		return
	}

	if err := Send(in.Phone, in.Email); err != nil {
		switch err {
		case ErrRateLimited:
			c.JSON(http.StatusTooManyRequests, gin.H{"detail": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to send OTP"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent"})
}

func handleVerify(c *gin.Context) {
	var in VerifyInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	if in.Phone == "" && in.Email == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "phone or email required"})
		return
	}

	if err := Verify(in.Phone, in.Email, in.Code); err != nil {
		switch err {
		case ErrInvalidCode:
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid or expired OTP"})
		case ErrTooManyAttempts:
			c.JSON(http.StatusTooManyRequests, gin.H{"detail": "Too many attempts, request a new OTP"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "Verification failed"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"verified": true})
}
