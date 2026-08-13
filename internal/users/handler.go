package users

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

type registerReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required"`
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type forgotPasswordReq struct {
	Email string `json:"email" binding:"required,email"`
}

type resetPasswordReq struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type updateMeReq struct {
	FullName         *string `json:"full_name"`
	BusinessCategory *string `json:"business_category"`
}

type changePasswordReq struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// meResponse adds the computed quality tier alongside the raw user record
// — computed, not stored, so it can't drift from QualityFlaggedAt.
type meResponse struct {
	User
	QualityStatus string `json:"quality_status"`
}

func withQualityStatus(u *User) meResponse {
	return meResponse{User: *u, QualityStatus: QualityStatus(u)}
}

type tokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

func RegisterRoutes(r *gin.Engine) {
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", handleRegister)
		authGroup.POST("/login", handleLogin)
		authGroup.POST("/refresh", handleRefresh)
		authGroup.POST("/forgot-password", handleForgotPassword)
		authGroup.POST("/reset-password", handleResetPassword)
	}

	userGroup := r.Group("/users")
	userGroup.Use(auth.AuthRequired())
	{
		userGroup.GET("/me", handleMe)
		userGroup.PATCH("/me", handleUpdateMe)
		userGroup.POST("/me/change-password", handleChangePassword)
	}
}

func handleRegister(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}

	user, err := CreateUser(database.DB, req.Email, req.Password, req.FullName)
	if err == ErrEmailTaken {
		c.JSON(http.StatusConflict, gin.H{"detail": "Email already registered"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func handleLogin(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}

	user, err := Authenticate(database.DB, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": err.Error()})
		return
	}

	accessToken, err := auth.CreateAccessToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to create token"})
		return
	}
	refreshToken, err := auth.CreateRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to create token"})
		return
	}

	c.JSON(http.StatusOK, tokenResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
	})
}

func handleRefresh(c *gin.Context) {
	var req refreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}

	userID, err := auth.ValidateToken(req.RefreshToken, auth.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid or expired refresh token"})
		return
	}

	user, err := GetByID(database.DB, userID)
	if err != nil || !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid or expired refresh token"})
		return
	}

	// Rotate: issue a fresh pair (access + refresh). Replay detection
	// via jti tracking comes in Phase 2.
	accessToken, _ := auth.CreateAccessToken(user.ID)
	refreshToken, _ := auth.CreateRefreshToken(user.ID)

	c.JSON(http.StatusOK, tokenResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
	})
}

func handleMe(c *gin.Context) {
	userID := auth.GetUserID(c)
	user, err := GetByID(database.DB, userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "User not found"})
		return
	}
	c.JSON(http.StatusOK, withQualityStatus(user))
}

func handleUpdateMe(c *gin.Context) {
	userID := auth.GetUserID(c)
	user, err := GetByID(database.DB, userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "User not found"})
		return
	}

	var req updateMeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}

	updated, err := UpdateProfile(database.DB, user, &UpdateProfileInput{
		FullName:         req.FullName,
		BusinessCategory: req.BusinessCategory,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, withQualityStatus(updated))
}

func handleChangePassword(c *gin.Context) {
	userID := auth.GetUserID(c)
	user, err := GetByID(database.DB, userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "User not found"})
		return
	}

	var req changePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}

	if err := ChangePassword(database.DB, user, req.CurrentPassword, req.NewPassword); err == ErrWrongCurrentPassword {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"detail": "Password updated"})
}

// handleForgotPassword — always returns 200 (with a generic message)
// regardless of whether the email exists, so an attacker can't enumerate
// registered addresses through this endpoint. When the email does exist
// we issue a short-lived JWT-encoded reset token and (in dev) log the
// full reset URL; a real deployment wires this to the email provider.
func handleForgotPassword(c *gin.Context) {
	var req forgotPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	generic := gin.H{"detail": "If that address is registered, a reset link has been sent."}

	user, err := GetByEmail(database.DB, req.Email)
	if err != nil {
		c.JSON(http.StatusOK, generic)
		return
	}
	token, err := auth.CreatePasswordResetToken(user.ID)
	if err != nil {
		c.JSON(http.StatusOK, generic)
		return
	}
	// Dev-only surface for the reset URL — production wires this to an
	// email sender in the notifications/webhooks layer.
	log.Printf("password reset (dev): http://localhost:5173/reset-password?token=%s", token)
	c.JSON(http.StatusOK, generic)
}

// handleResetPassword accepts the reset token issued above plus a new
// password and rotates it. On success the caller can log in normally.
func handleResetPassword(c *gin.Context) {
	var req resetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	userID, err := auth.ValidatePasswordResetToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid or expired reset token"})
		return
	}
	user, err := GetByID(database.DB, userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid or expired reset token"})
		return
	}
	hashed, err := HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	user.HashedPassword = hashed
	if err := database.DB.Save(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"detail": "Password reset"})
}
