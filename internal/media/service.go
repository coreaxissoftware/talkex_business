package media

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/config"
)

// safeMimeTypes is the allowlist of MIME types the server will accept
// on upload AND echo back on serve. Anything else is stored as
// application/octet-stream and served with Content-Disposition: attachment
// so the browser never renders it inline on our origin. Uploading an
// HTML file with Content-Type text/html was the classic same-origin XSS
// vector before this guard existed.
var safeMimeTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"image/webp":      true,
	"image/svg+xml":   false, // SVG can carry <script>; never inline.
	"video/mp4":       true,
	"video/webm":      true,
	"video/quicktime": true,
	"audio/mpeg":      true,
	"audio/ogg":       true,
	"audio/wav":       true,
	"application/pdf": true,
	"text/plain":      true,
	"text/csv":        true,
}

// safeExtensions are the file-name extensions accepted for the stored
// filename. Everything else is coerced to .bin so nothing on the server
// can accidentally interpret the file (e.g. a .php served by a static
// host mounted elsewhere).
var safeExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	".webp": true, ".mp4": true, ".webm": true, ".mov": true,
	".mp3": true, ".ogg": true, ".wav": true, ".pdf": true,
	".txt": true, ".csv": true,
}

// SafeMimeAndDisposition returns the MIME type and Content-Disposition
// value the /media/files handler should send for the given stored row.
// Non-inline types force a download; SVG in particular is always
// downloaded so its embedded scripts never run on our origin.
func SafeMimeAndDisposition(m *Media) (mime, disposition string) {
	mime = strings.ToLower(m.MimeType)
	if !safeMimeTypes[mime] {
		mime = "application/octet-stream"
		// Escape quotes in the filename per RFC 6266.
		safe := strings.ReplaceAll(m.OriginalName, `"`, `\"`)
		return mime, fmt.Sprintf(`attachment; filename="%s"`, safe)
	}
	safe := strings.ReplaceAll(m.OriginalName, `"`, `\"`)
	return mime, fmt.Sprintf(`inline; filename="%s"`, safe)
}

// signatureTTL is how long a signed media URL stays valid. Kept short
// so a leaked link can't be replayed for long; extended by re-listing
// (which mints fresh signatures on every /media response).
const signatureTTL = 30 * time.Minute

// SignedURLFor returns the served URL for a stored filename, carrying
// an HMAC that binds filename+expiry to the JWT secret. This lets the
// browser use plain `<img src>` without an Authorization header while
// still preventing anonymous cross-tenant enumeration.
func SignedURLFor(filename string) string {
	exp := time.Now().Add(signatureTTL).Unix()
	mac := hmac.New(sha256.New, []byte(config.Get().JWTSecret))
	fmt.Fprintf(mac, "%s|%d", filename, exp)
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("/media/files/%s?exp=%d&sig=%s", filename, exp, sig)
}

// VerifySignature checks the exp+sig pair against the JWT secret and
// rejects expired signatures. Returns true only when the signature
// matches AND the expiry is still in the future.
func VerifySignature(filename, expStr, sig string) bool {
	if filename == "" || sig == "" || expStr == "" {
		return false
	}
	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return false
	}
	if time.Now().Unix() > exp {
		return false
	}
	mac := hmac.New(sha256.New, []byte(config.Get().JWTSecret))
	fmt.Fprintf(mac, "%s|%d", filename, exp)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sig))
}

var (
	ErrMediaNotFound = errors.New("media not found")
	UploadDir        = "uploads"
)

func init() {
	os.MkdirAll(UploadDir, 0o755)
}

func List(db *gorm.DB, ownerID string) ([]Media, error) {
	var out []Media
	err := db.Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&out).Error
	if err == nil {
		stampSignedURLs(out)
	}
	return out, err
}

func GetByID(db *gorm.DB, ownerID, id string) (*Media, error) {
	var m Media
	err := db.Where("id = ? AND owner_id = ?", id, ownerID).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMediaNotFound
	}
	return &m, err
}

// GetByFilename resolves the stored (UUID) filename to a Media row,
// scoped to the caller's owner_id. Used by the auth-gated file server.
func GetByFilename(db *gorm.DB, ownerID, filename string) (*Media, error) {
	var m Media
	err := db.Where("filename = ? AND owner_id = ?", filename, ownerID).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMediaNotFound
	}
	return &m, err
}

func Upload(db *gorm.DB, ownerID, originalName, mimeType string, size int64, reader io.Reader) (*Media, error) {
	// Sniff the first 512 bytes so we trust our detector over the
	// client-supplied Content-Type header (attackers can send anything).
	sniff := make([]byte, 512)
	n, _ := io.ReadFull(reader, sniff)
	sniff = sniff[:n]
	detectedMime := http.DetectContentType(sniff)
	trusted := strings.ToLower(strings.SplitN(detectedMime, ";", 2)[0])
	claimed := strings.ToLower(strings.TrimSpace(mimeType))
	// text/csv and text/plain are indistinguishable to the detector;
	// honor the client claim only if it's on the allowlist.
	if (claimed == "text/csv" || claimed == "text/plain") && strings.HasPrefix(trusted, "text/") {
		trusted = claimed
	}
	if !safeMimeTypes[trusted] {
		return nil, fmt.Errorf("mime type %q not allowed", trusted)
	}

	// Extension allow-list — coerce to .bin if unknown/unsafe.
	ext := strings.ToLower(filepath.Ext(filepath.Base(originalName)))
	if len(ext) > 10 || !safeExtensions[ext] {
		ext = ".bin"
	}
	// Sanitize the original name we display back — strip path bits.
	displayName := filepath.Base(originalName)
	if len(displayName) > 255 {
		displayName = displayName[:255]
	}

	storedName := uuid.New().String() + ext
	destPath := filepath.Join(UploadDir, storedName)

	f, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	// Write the sniff buffer back first, then stream the rest.
	if _, err := io.Copy(f, io.MultiReader(bytes.NewReader(sniff), reader)); err != nil {
		os.Remove(destPath)
		return nil, fmt.Errorf("write file: %w", err)
	}

	m := &Media{
		OwnerID:      ownerID,
		Filename:     storedName,
		OriginalName: displayName,
		MimeType:     trusted, // Store the sniffed, allow-listed MIME.
		Size:         size,
		// URL is regenerated on every read with a fresh signature so
		// the stored value doesn't stale. Populate for the response
		// of this single POST.
		URL: SignedURLFor(storedName),
	}
	if err := db.Create(m).Error; err != nil {
		os.Remove(destPath)
		return nil, err
	}
	return m, nil
}

// stampSignedURLs mints a fresh short-lived signature on every read so
// consumers don't have to know about the signature scheme.
func stampSignedURLs(items []Media) {
	for i := range items {
		items[i].URL = SignedURLFor(items[i].Filename)
	}
}

func Delete(db *gorm.DB, m *Media) error {
	os.Remove(filepath.Join(UploadDir, m.Filename))
	return db.Delete(m).Error
}
