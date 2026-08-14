package compliance

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/compliance")
	g.Use(auth.AuthRequired())
	{
		g.GET("/stats", handleStats)

		// Consent management
		g.GET("/consents", handleListAllConsents)
		g.GET("/consents/:contactId", handleListConsents)
		g.POST("/consents", handleRecordConsent)
		g.POST("/consents/:contactId/revoke-all", handleRevokeAll)

		// DSAR
		g.GET("/dsars", handleListDSARs)
		g.POST("/dsars", handleCreateDSAR)
		g.POST("/dsars/:id/process", handleProcessDSAR)
		g.POST("/dsars/:id/complete", handleCompleteDSAR)
		g.POST("/dsars/:id/reject", handleRejectDSAR)

		// Processing records
		g.GET("/processing", handleListProcessing)
	}
}

func handleStats(c *gin.Context) {
	stats, err := GetStats(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func handleListAllConsents(c *gin.Context) {
	items, err := ListAllConsents(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func handleListConsents(c *gin.Context) {
	items, err := ListConsents(database.DB, auth.GetUserID(c), c.Param("contactId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func handleRecordConsent(c *gin.Context) {
	var in ConsentInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	rec, err := RecordConsent(database.DB, auth.GetUserID(c), &in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, rec)
}

func handleRevokeAll(c *gin.Context) {
	count, err := RevokeAllConsents(database.DB, auth.GetUserID(c), c.Param("contactId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"revoked": count})
}

func handleListDSARs(c *gin.Context) {
	items, err := ListDSARs(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func handleCreateDSAR(c *gin.Context) {
	var in DSARInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	req, err := CreateDSAR(database.DB, auth.GetUserID(c), &in)
	if err != nil {
		if err == ErrInvalidType {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "Invalid type. Use: access, erasure, correction, portability"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, req)
}

func handleProcessDSAR(c *gin.Context) {
	req, err := ProcessDSAR(database.DB, auth.GetUserID(c), c.Param("id"))
	if err != nil {
		if err == ErrDSARNotFound {
			c.JSON(http.StatusNotFound, gin.H{"detail": "DSAR request not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, req)
}

func handleCompleteDSAR(c *gin.Context) {
	var body struct {
		Response string `json:"response"`
	}
	c.ShouldBindJSON(&body)
	req, err := CompleteDSAR(database.DB, auth.GetUserID(c), c.Param("id"), body.Response)
	if err != nil {
		if err == ErrDSARNotFound {
			c.JSON(http.StatusNotFound, gin.H{"detail": "DSAR request not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, req)
}

func handleRejectDSAR(c *gin.Context) {
	var body struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&body)
	req, err := RejectDSAR(database.DB, auth.GetUserID(c), c.Param("id"), body.Reason)
	if err != nil {
		if err == ErrDSARNotFound {
			c.JSON(http.StatusNotFound, gin.H{"detail": "DSAR request not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, req)
}

func handleListProcessing(c *gin.Context) {
	items, err := ListProcessingRecords(database.DB, auth.GetUserID(c), 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, items)
}
