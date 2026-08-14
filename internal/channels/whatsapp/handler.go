package whatsapp

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/channels/whatsapp/onboarding")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleGetOnboarding)
		g.POST("/start", handleStartOnboarding)
		g.PUT("/business-info", handleBusinessInfo)
		g.PUT("/verification", handleVerification)
		g.PUT("/phone-registration", handlePhoneRegistration)
		g.PUT("/display-name", handleDisplayName)
	}
}

func handleGetOnboarding(c *gin.Context) {
	ownerID := auth.GetUserID(c)
	o, err := GetOnboarding(database.DB, ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	if o == nil {
		c.JSON(http.StatusOK, gin.H{"status": "not_started"})
		return
	}
	c.JSON(http.StatusOK, o)
}

func handleStartOnboarding(c *gin.Context) {
	ownerID := auth.GetUserID(c)
	o, err := StartOnboarding(database.DB, ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, o)
}

func handleBusinessInfo(c *gin.Context) {
	ownerID := auth.GetUserID(c)
	o, err := GetOnboarding(database.DB, ownerID)
	if err != nil || o == nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Start onboarding first"})
		return
	}
	var req BusinessInfoInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	if err := SaveBusinessInfo(database.DB, o, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, o)
}

func handleVerification(c *gin.Context) {
	ownerID := auth.GetUserID(c)
	o, err := GetOnboarding(database.DB, ownerID)
	if err != nil || o == nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Start onboarding first"})
		return
	}
	var req VerificationInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	if err := SaveVerification(database.DB, o, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, o)
}

func handlePhoneRegistration(c *gin.Context) {
	ownerID := auth.GetUserID(c)
	o, err := GetOnboarding(database.DB, ownerID)
	if err != nil || o == nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Start onboarding first"})
		return
	}
	var req PhoneRegistrationInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	if err := SavePhoneRegistration(database.DB, o, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, o)
}

func handleDisplayName(c *gin.Context) {
	ownerID := auth.GetUserID(c)
	o, err := GetOnboarding(database.DB, ownerID)
	if err != nil || o == nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Start onboarding first"})
		return
	}
	var req DisplayNameInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	if err := SaveDisplayName(database.DB, o, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, o)
}
