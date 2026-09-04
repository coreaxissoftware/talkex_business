package deals

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/apihelpers"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/deals")
	g.Use(auth.AuthRequired())
	{
		g.GET("/pipelines", handleListPipelines)
		g.GET("/pipelines/default", handleDefaultPipeline)
		g.GET("/pipelines/:id/kanban", handleKanban)
		g.POST("", handleCreate)
		g.POST("/:id/move", handleMove)
	}
}

func handleListPipelines(c *gin.Context) {
	items, err := ListPipelines(database.DB, auth.GetUserID(c))
	if err != nil {
		apihelpers.ServerError(c, err, "internal")
		return
	}
	c.JSON(http.StatusOK, items)
}

func handleDefaultPipeline(c *gin.Context) {
	p, err := EnsureDefaultPipeline(database.DB, auth.GetUserID(c))
	if err != nil {
		apihelpers.ServerError(c, err, "internal")
		return
	}
	c.JSON(http.StatusOK, p)
}

func handleKanban(c *gin.Context) {
	cols, err := Kanban(database.DB, auth.GetUserID(c), c.Param("id"))
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Pipeline not found"})
		return
	}
	if err != nil {
		apihelpers.ServerError(c, err, "internal")
		return
	}
	c.JSON(http.StatusOK, cols)
}

func handleCreate(c *gin.Context) {
	var d Deal
	if err := c.ShouldBindJSON(&d); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	if err := CreateDeal(database.DB, auth.GetUserID(c), &d); err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"detail": "Pipeline not found"})
		case errors.Is(err, ErrInvalidStage):
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Stage not part of the pipeline"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, d)
}

func handleMove(c *gin.Context) {
	var body struct {
		Stage string `json:"stage" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	d, err := MoveDeal(database.DB, auth.GetUserID(c), c.Param("id"), body.Stage)
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Deal not found"})
		return
	}
	if errors.Is(err, ErrInvalidStage) {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Stage not part of the pipeline"})
		return
	}
	if err != nil {
		apihelpers.ServerError(c, err, "internal")
		return
	}
	c.JSON(http.StatusOK, d)
}
