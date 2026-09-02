package whitelabel

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// defaultBranding is what the frontend renders when no override exists.
// Matches the marketing site tokens so first-run installs still look
// intentional rather than half-styled.
func defaultBranding() *Branding {
	return &Branding{
		BrandName:    "TalkEx Business",
		Tagline:      "One inbox. Every messaging channel.",
		PrimaryColor: "#0EA5A0",
		AccentColor:  "#F97066",
		FromEmail:    "no-reply@business.talkex.in",
	}
}

func RegisterRoutes(r *gin.Engine) {
	// Public — the login / signup page hits this BEFORE the user has
	// a token, so it can render the reseller's logo. Looked up by
	// Host header (custom domain) — falls back to defaults when no
	// row matches.
	r.GET("/branding", handlePublicBranding)

	g := r.Group("/branding")
	g.Use(auth.AuthRequired())
	{
		g.GET("/mine", handleGet)
		g.PUT("/mine", handleUpdate)
	}
}

// handlePublicBranding — Host-header lookup. Returns the branding row
// whose CustomDomain matches c.Request.Host (case-insensitive, port
// stripped). Anonymous, cached briefly at the edge.
func handlePublicBranding(c *gin.Context) {
	host := c.Request.Host
	// Strip :port so localhost:5173 matches "localhost".
	if i := indexByte(host, ':'); i >= 0 {
		host = host[:i]
	}

	var b Branding
	err := database.DB.Where("custom_domain = ?", host).First(&b).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.Header("Cache-Control", "public, max-age=60")
		c.JSON(http.StatusOK, defaultBranding())
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.Header("Cache-Control", "public, max-age=60")
	c.JSON(http.StatusOK, &b)
}

func handleGet(c *gin.Context) {
	ownerID := auth.GetUserID(c)
	var b Branding
	err := database.DB.Where("owner_id = ?", ownerID).First(&b).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Return the default so the frontend form fields have something
		// sensible to populate.
		def := defaultBranding()
		def.OwnerID = ownerID
		c.JSON(http.StatusOK, def)
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, &b)
}

func handleUpdate(c *gin.Context) {
	ownerID := auth.GetUserID(c)
	var in Branding
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	// Normalise + guard: colours must be #-prefixed hex, custom domain
	// stripped of scheme.
	in.PrimaryColor = normHex(in.PrimaryColor)
	in.AccentColor = normHex(in.AccentColor)
	in.CustomDomain = stripScheme(in.CustomDomain)

	var existing Branding
	err := database.DB.Where("owner_id = ?", ownerID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		in.OwnerID = ownerID
		if err := database.DB.Create(&in).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
			return
		}
		c.JSON(http.StatusCreated, &in)
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}

	// Preserve OwnerID + ID + timestamps; overwrite the rest.
	in.ID = existing.ID
	in.OwnerID = ownerID
	in.CreatedAt = existing.CreatedAt
	if err := database.DB.Save(&in).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, &in)
}

// normHex prefixes a bare "0EA5A0" with '#' and lowercases. Empty stays empty.
func normHex(s string) string {
	if s == "" {
		return ""
	}
	if s[0] != '#' {
		s = "#" + s
	}
	// Lowercase in place.
	out := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		out[i] = c
	}
	return string(out)
}

// stripScheme trims "https://" / "http://" from a domain and drops a
// trailing slash.
func stripScheme(s string) string {
	for _, p := range []string{"https://", "http://"} {
		if len(s) > len(p) && s[:len(p)] == p {
			s = s[len(p):]
			break
		}
	}
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}
