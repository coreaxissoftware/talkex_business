package media

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
		URL:          "/media/files/" + storedName,
	}
	if err := db.Create(m).Error; err != nil {
		os.Remove(destPath)
		return nil, err
	}
	return m, nil
}

func Delete(db *gorm.DB, m *Media) error {
	os.Remove(filepath.Join(UploadDir, m.Filename))
	return db.Delete(m).Error
}
