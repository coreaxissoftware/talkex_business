package otp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes adds OTP endpoints under /auth/otp.
func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/auth/otp")
	{
		g.POST("/send", handleSend)
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
