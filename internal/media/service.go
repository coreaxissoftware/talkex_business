package media

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/config"
)

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
	ext := filepath.Ext(originalName)
	storedName := uuid.New().String() + ext
	destPath := filepath.Join(UploadDir, storedName)

	f, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		os.Remove(destPath)
		return nil, fmt.Errorf("write file: %w", err)
	}

	m := &Media{
		OwnerID:      ownerID,
		Filename:     storedName,
		OriginalName: originalName,
		MimeType:     mimeType,
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
