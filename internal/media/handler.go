package media

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/media")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleList)
		g.POST("/upload", handleUpload)
		g.DELETE("/:id", handleDelete)
	}

	r.Static("/media/files", filepath.Join(UploadDir))
}

func handleList(c *gin.Context) {
	items, err := List(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, items)
}

const maxUploadSize = 10 << 20 // 10 MB

func handleUpload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Missing or oversized file (max 10 MB)"})
		return
	}
	defer file.Close()

	m, err := Upload(
		database.DB,
		auth.GetUserID(c),
		header.Filename,
		header.Header.Get("Content-Type"),
		header.Size,
		file,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Upload failed"})
		return
	}
	c.JSON(http.StatusCreated, m)
}

func handleDelete(c *gin.Context) {
	m, err := GetByID(database.DB, auth.GetUserID(c), c.Param("id"))
	if err == ErrMediaNotFound {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Media not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	if err := Delete(database.DB, m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.Status(http.StatusNoContent)
}
