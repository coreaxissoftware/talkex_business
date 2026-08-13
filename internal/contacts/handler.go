package contacts

import (
	"encoding/csv"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/contacts")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleList)
		g.POST("", handleCreate)
		g.GET("/:id", handleGet)
		g.PATCH("/:id", handleUpdate)
		g.DELETE("/:id", handleDelete)
		g.POST("/import-csv", handleImportCSV)
	}
}

func handleList(c *gin.Context) {
	contacts, err := List(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, contacts)
}

func handleCreate(c *gin.Context) {
	var in CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	contact, err := Create(database.DB, auth.GetUserID(c), &in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, contact)
}

func getOwnedOrAbort(c *gin.Context) *Contact {
	contact, err := GetByID(database.DB, auth.GetUserID(c), c.Param("id"))
	if err == ErrContactNotFound {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Contact not found"})
		return nil
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return nil
	}
	return contact
}

func handleGet(c *gin.Context) {
	if contact := getOwnedOrAbort(c); contact != nil {
		c.JSON(http.StatusOK, contact)
	}
}

func handleUpdate(c *gin.Context) {
	contact := getOwnedOrAbort(c)
	if contact == nil {
		return
	}
	var in UpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	updated, err := Update(database.DB, contact, &in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func handleDelete(c *gin.Context) {
	contact := getOwnedOrAbort(c)
	if contact == nil {
		return
	}
	if err := Delete(database.DB, contact); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.Status(http.StatusNoContent)
}

func handleImportCSV(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Missing file field"})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "Cannot read CSV header"})
		return
	}

	colMap := map[string]int{}
	for i, h := range header {
		colMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	phoneCol, ok := colMap["phone"]
	if !ok {
		if idx, ok2 := colMap["phone_number"]; ok2 {
			phoneCol = idx
		} else {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "CSV must have a 'phone' or 'phone_number' column"})
			return
		}
	}

	ownerID := auth.GetUserID(c)
	var created, skipped, failed int
	var errors []string

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			failed++
			continue
		}

		phone := strings.TrimSpace(row[phoneCol])
		if phone == "" {
			skipped++
			continue
		}

		in := &CreateInput{PhoneNumber: phone}

		if idx, ok := colMap["name"]; ok && idx < len(row) {
			val := strings.TrimSpace(row[idx])
			if val != "" {
				in.Name = &val
			}
		}
		if idx, ok := colMap["email"]; ok && idx < len(row) {
			val := strings.TrimSpace(row[idx])
			if val != "" {
				in.Email = &val
			}
		}
		if idx, ok := colMap["tags"]; ok && idx < len(row) {
			val := strings.TrimSpace(row[idx])
			if val != "" {
				parts := strings.Split(val, ";")
				tags := make([]string, 0, len(parts))
				for _, p := range parts {
					t := strings.TrimSpace(p)
					if t != "" {
						tags = append(tags, t)
					}
				}
				in.Tags = tags
			}
		}

		_, createErr := Create(database.DB, ownerID, in)
		if createErr != nil {
			failed++
			errors = append(errors, phone+": "+createErr.Error())
		} else {
			created++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"created": created,
		"skipped": skipped,
		"failed":  failed,
		"errors":  errors,
	})
}
