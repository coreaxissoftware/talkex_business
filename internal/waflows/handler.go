package waflows

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/waflows")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleList)
		g.POST("", handleCreate)
		g.GET("/:id", handleGet)
		g.PUT("/:id", handleUpdate)
		g.POST("/:id/publish", handlePublish)
		g.GET("/:id/responses", handleListResponses)
	}
	// Meta pushes end-user submissions here. Public — no auth. In prod,
	// Meta signs the request with the app secret; verification lives in
	// internal/channels/whatsapp (added later — for now we accept the
	// payload and log). The owner_id path segment scopes the row.
	r.POST("/waflows/inbound/:owner", handleInbound)
}

func handleList(c *gin.Context) {
	ownerID := auth.GetUserID(c)
	var items []WAFlow
	if err := database.DB.Where("owner_id = ?", ownerID).
		Order("created_at DESC").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func handleGet(c *gin.Context) {
	var f WAFlow
	err := database.DB.Where("id = ? AND owner_id = ?", c.Param("id"), auth.GetUserID(c)).First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Flow not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, f)
}

type createInput struct {
	Name     string          `json:"name" binding:"required"`
	Category string          `json:"category"`
	FlowJSON json.RawMessage `json:"flow_json" binding:"required"`
	Endpoint string          `json:"endpoint"`
}

func handleCreate(c *gin.Context) {
	var in createInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	if in.Category == "" {
		in.Category = CategoryOther
	}
	// Validate the FlowJSON is at least parseable JSON with a routing_model.
	var probe struct {
		RoutingModel map[string]interface{} `json:"routing_model"`
		Screens      []map[string]interface{} `json:"screens"`
	}
	if err := json.Unmarshal(in.FlowJSON, &probe); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "flow_json must be an object with routing_model + screens"})
		return
	}
	if len(probe.Screens) == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "flow_json must define at least one screen"})
		return
	}

	f := WAFlow{
		OwnerID:  auth.GetUserID(c),
		Name:     in.Name,
		Category: in.Category,
		Status:   StatusDraft,
		Version:  1,
		FlowJSON: datatypes.JSON(in.FlowJSON),
		Endpoint: in.Endpoint,
	}
	if err := database.DB.Create(&f).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, f)
}

type updateInput struct {
	Name     *string          `json:"name"`
	Category *string          `json:"category"`
	FlowJSON json.RawMessage  `json:"flow_json"`
	Endpoint *string          `json:"endpoint"`
}

func handleUpdate(c *gin.Context) {
	var f WAFlow
	err := database.DB.Where("id = ? AND owner_id = ?", c.Param("id"), auth.GetUserID(c)).First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Flow not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	// Editing a published flow forces a version bump so previous sends
	// stay pinned to the historic definition.
	if f.Status == StatusPublished {
		f.Version++
		f.Status = StatusDraft
	}

	var in updateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	if in.Name != nil {
		f.Name = *in.Name
	}
	if in.Category != nil {
		f.Category = *in.Category
	}
	if len(in.FlowJSON) > 0 {
		f.FlowJSON = datatypes.JSON(in.FlowJSON)
	}
	if in.Endpoint != nil {
		f.Endpoint = *in.Endpoint
	}
	if err := database.DB.Save(&f).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, f)
}

// handlePublish would push to Meta's Flow API in production. In dev
// (or without Meta credentials) it just stamps the row.
func handlePublish(c *gin.Context) {
	var f WAFlow
	err := database.DB.Where("id = ? AND owner_id = ?", c.Param("id"), auth.GetUserID(c)).First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Flow not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	// Deferred: Meta Graph API call. For now stamp status + a synthetic
	// FlowID so the messaging engine can reference it.
	now := time.Now()
	f.Status = StatusPublished
	f.PublishedAt = &now
	if f.MetaFlowID == "" {
		f.MetaFlowID = "sim-" + f.ID
	}
	if err := database.DB.Save(&f).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, f)
}

func handleListResponses(c *gin.Context) {
	ownerID := auth.GetUserID(c)
	var out []FlowResponse
	if err := database.DB.Where("owner_id = ? AND flow_id = ?", ownerID, c.Param("id")).
		Order("created_at DESC").Limit(200).Find(&out).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, out)
}

// handleInbound receives an end-user submission from Meta and stores
// it under the owner scoped by the URL segment. Body shape mirrors
// Meta's Flows data-collection webhook.
func handleInbound(c *gin.Context) {
	ownerID := c.Param("owner")
	if ownerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "owner segment required"})
		return
	}
	var payload struct {
		FlowID    string          `json:"flow_id"`
		ScreenID  string          `json:"screen_id"`
		ContactID string          `json:"contact_id"`
		Data      json.RawMessage `json:"data"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	if payload.FlowID == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "flow_id required"})
		return
	}
	resp := FlowResponse{
		OwnerID:   ownerID,
		FlowID:    payload.FlowID,
		ContactID: payload.ContactID,
		ScreenID:  payload.ScreenID,
		Data:      datatypes.JSON(payload.Data),
	}
	if err := database.DB.Create(&resp).Error; err != nil {
		log.Printf("waflows: inbound persist failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "received"})
}
