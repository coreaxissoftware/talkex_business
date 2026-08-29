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

	// File serving is public-with-signature: the URL carries an HMAC
	// over (filename+expiry) signed with the JWT secret. Lets browsers
	// use plain `<img src>` (no Authorization header) while still
	// preventing anonymous cross-tenant enumeration. Signatures expire
	// in signatureTTL (30 min) and are minted fresh on every /media
	// list, so no long-lived leak is possible.
	r.GET("/media/files/:filename", handleServeFile)
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

// handleServeFile streams the raw bytes of an uploaded file after
// verifying the signed URL. The URL carries an HMAC over (filename,
// expiry) signed with the JWT secret; without a valid, non-expired
// signature the request is rejected. Signatures are minted by the
// /media list endpoint and last signatureTTL.
func handleServeFile(c *gin.Context) {
	filename := c.Param("filename")
	// Path-traversal defense — normalize before doing anything else.
	if filename == "" || filename != filepath.Base(filename) {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid filename"})
		return
	}

	if !VerifySignature(filename, c.Query("exp"), c.Query("sig")) {
		c.JSON(http.StatusForbidden, gin.H{"detail": "Invalid or expired signature"})
		return
	}

	// Signature valid — resolve mime type from the DB row for the
	// Content-Type header, but do NOT enforce owner_id here (the
	// signature itself is the authorization).
	var m Media
	if err := database.DB.Where("filename = ?", filename).First(&m).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Media not found"})
		return
	}

	path := filepath.Join(UploadDir, m.Filename)
	c.Header("Content-Type", m.MimeType)
	c.File(path)
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
