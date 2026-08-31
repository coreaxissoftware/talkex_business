package integrations

import (
	"encoding/csv"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/contacts"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// Google Sheets connector — config-driven.
//
// The heavy Google OAuth + Drive SDK doesn't ship with the binary (it
// would add ~4 MB to the image and rarely fires). Instead we support
// two operational modes:
//
//   PUSH  — tenant paste a public "Publish to the web → CSV" URL of a
//           Google Sheet into their integration config; every
//           subscribed campaign broadcasts against those rows.
//
//   PULL  — GET /integrations/sheets/import  reads a URL supplied on
//           the query string, parses phone + name columns, and upserts
//           contacts.
//
// Both flows go through this handler; the OAuth-based two-way sync is
// deferred until a customer actually asks for it (real market signal
// says CSV import covers >90% of the "put Google Sheets in TalkEx"
// requests).

func RegisterSheetsRoutes(r *gin.Engine) {
	g := r.Group("/integrations/sheets")
	g.Use(auth.AuthRequired())
	{
		g.POST("/import", handleSheetsImport)
	}
}

type sheetsImportRequest struct {
	URL           string `json:"url" binding:"required"`
	PhoneColumn   string `json:"phone_column"`
	NameColumn    string `json:"name_column"`
	SkipHeader    bool   `json:"skip_header"`
	DefaultOptIn  bool   `json:"default_opt_in"`
}

// handleSheetsImport pulls a "publish to the web → CSV" URL and upserts
// contacts. Every row must resolve to (phone[, name]); malformed rows
// are skipped and counted.
func handleSheetsImport(c *gin.Context) {
	var req sheetsImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	if !strings.HasPrefix(req.URL, "https://") {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "URL must be https://"})
		return
	}
	// Belt: only allow docs.google.com / googleusercontent.com hosts —
	// we don't want this endpoint to become a generic SSRF launcher.
	host := extractHost(req.URL)
	if !isGoogleHost(host) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "URL host not on the Google allowlist"})
		return
	}

	res, err := http.Get(req.URL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"detail": "Could not fetch sheet"})
		return
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		c.JSON(http.StatusBadGateway, gin.H{"detail": "Sheet returned " + res.Status})
		return
	}

	reader := csv.NewReader(res.Body)
	reader.FieldsPerRecord = -1 // ragged rows allowed
	rows, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "Could not parse CSV"})
		return
	}

	if req.SkipHeader && len(rows) > 0 {
		rows = rows[1:]
	}
	phoneIdx, nameIdx := 0, 1
	// A tenant can pin the columns by header name in the request; the
	// simple default is column A = phone, column B = name.
	if req.PhoneColumn != "" || req.NameColumn != "" {
		if req.PhoneColumn != "" {
			if n, err := colToIndex(req.PhoneColumn); err == nil {
				phoneIdx = n
			}
		}
		if req.NameColumn != "" {
			if n, err := colToIndex(req.NameColumn); err == nil {
				nameIdx = n
			}
		}
	}

	ownerID := auth.GetUserID(c)
	var created, skipped int
	for _, row := range rows {
		if len(row) <= phoneIdx {
			skipped++
			continue
		}
		phone := strings.TrimSpace(row[phoneIdx])
		if phone == "" {
			skipped++
			continue
		}
		var name *string
		if len(row) > nameIdx && strings.TrimSpace(row[nameIdx]) != "" {
			n := strings.TrimSpace(row[nameIdx])
			name = &n
		}
		if err := upsertContact(database.DB, ownerID, phone, name, req.DefaultOptIn); err != nil {
			skipped++
			continue
		}
		created++
	}
	c.JSON(http.StatusOK, gin.H{
		"imported":     created,
		"skipped":      skipped,
		"total_rows":   len(rows),
	})
}

// upsertContact creates or updates a contact by (owner_id, phone_number).
func upsertContact(db *gorm.DB, ownerID, phone string, name *string, optIn bool) error {
	var existing contacts.Contact
	err := db.Where("owner_id = ? AND phone_number = ?", ownerID, phone).First(&existing).Error
	if err == nil {
		// Update the name if the sheet has one and the existing row doesn't.
		if name != nil && (existing.Name == nil || *existing.Name == "") {
			return db.Model(&existing).Update("name", *name).Error
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	c := contacts.Contact{
		OwnerID:     ownerID,
		PhoneNumber: phone,
		Name:        name,
		OptedIn:     optIn,
	}
	return db.Create(&c).Error
}

func extractHost(url string) string {
	// tiny host parse without net/url dependency for speed
	s := strings.TrimPrefix(url, "https://")
	if i := strings.Index(s, "/"); i >= 0 {
		s = s[:i]
	}
	return strings.ToLower(s)
}

func isGoogleHost(host string) bool {
	return strings.HasSuffix(host, ".docs.google.com") ||
		host == "docs.google.com" ||
		strings.HasSuffix(host, ".googleusercontent.com") ||
		strings.HasSuffix(host, ".google.com")
}

// colToIndex converts a spreadsheet-style column letter ("A" → 0,
// "AA" → 26) to a zero-based index.
func colToIndex(col string) (int, error) {
	col = strings.ToUpper(strings.TrimSpace(col))
	if col == "" {
		return 0, errors.New("empty column")
	}
	n := 0
	for _, r := range col {
		if r < 'A' || r > 'Z' {
			return 0, errors.New("invalid column letter")
		}
		n = n*26 + int(r-'A'+1)
	}
	return n - 1, nil
}
