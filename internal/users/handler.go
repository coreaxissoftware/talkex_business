package users

import (
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
	}

	userGroup := r.Group("/users")
	userGroup.Use(auth.AuthRequired())
	{
		userGroup.GET("/me", handleMe)
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
	c.JSON(http.StatusOK, user)
}
