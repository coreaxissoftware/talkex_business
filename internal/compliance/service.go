package compliance

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrConsentNotFound = errors.New("consent record not found")
	ErrDSARNotFound    = errors.New("DSAR request not found")
	ErrInvalidType     = errors.New("invalid request type")
)

// ---- Consent Management ----

// RecordConsent creates or updates a consent record for a contact+purpose+channel combo.
func RecordConsent(db *gorm.DB, ownerID string, input *ConsentInput) (*ConsentRecord, error) {
	var existing ConsentRecord
	err := db.Where("owner_id = ? AND contact_id = ? AND purpose = ? AND channel = ?",
		ownerID, input.ContactID, input.Purpose, input.Channel).First(&existing).Error

	now := time.Now()
	if err == nil {
		// Update existing
		existing.ConsentGiven = input.ConsentGiven
		if input.ConsentGiven {
			existing.ConsentedAt = &now
			existing.RevokedAt = nil
		} else {
			existing.RevokedAt = &now
		}
		existing.Source = input.Source
		existing.IPAddress = input.IPAddress
		if err := db.Save(&existing).Error; err != nil {
			return nil, err
		}
		return &existing, nil
	}

	// Create new
	rec := &ConsentRecord{
		OwnerID:      ownerID,
		ContactID:    input.ContactID,
		Purpose:      input.Purpose,
		Channel:      input.Channel,
		ConsentGiven: input.ConsentGiven,
		Source:       input.Source,
		IPAddress:    input.IPAddress,
	}
	if input.ConsentGiven {
		rec.ConsentedAt = &now
	}
	if err := db.Create(rec).Error; err != nil {
		return nil, err
	}
	return rec, nil
}

type ConsentInput struct {
	ContactID    string  `json:"contact_id" binding:"required"`
	Purpose      string  `json:"purpose" binding:"required"`
	Channel      string  `json:"channel" binding:"required"`
	ConsentGiven bool    `json:"consent_given"`
	Source       string  `json:"source" binding:"required"`
	IPAddress    *string `json:"ip_address"`
}

// ListConsents returns all consent records for a contact.
func ListConsents(db *gorm.DB, ownerID, contactID string) ([]ConsentRecord, error) {
	var out []ConsentRecord
	err := db.Where("owner_id = ? AND contact_id = ?", ownerID, contactID).
		Order("created_at DESC").Find(&out).Error
	return out, err
}

// ListAllConsents returns all consent records for an owner (for audit).
func ListAllConsents(db *gorm.DB, ownerID string) ([]ConsentRecord, error) {
	var out []ConsentRecord
	err := db.Where("owner_id = ?", ownerID).
		Order("created_at DESC").Find(&out).Error
	return out, err
}

// RevokeAllConsents revokes all active consents for a contact (right to withdraw).
func RevokeAllConsents(db *gorm.DB, ownerID, contactID string) (int, error) {
	now := time.Now()
	result := db.Model(&ConsentRecord{}).
		Where("owner_id = ? AND contact_id = ? AND consent_given = ?", ownerID, contactID, true).
		Updates(map[string]interface{}{
			"consent_given": false,
			"revoked_at":    now,
		})
	return int(result.RowsAffected), result.Error
}

// ---- DSAR Requests ----

// CreateDSAR creates a new Data Subject Access Request.
func CreateDSAR(db *gorm.DB, ownerID string, input *DSARInput) (*DSARRequest, error) {
	validTypes := map[string]bool{"access": true, "erasure": true, "correction": true, "portability": true}
	if !validTypes[input.Type] {
		return nil, ErrInvalidType
	}

	req := &DSARRequest{
		OwnerID:   ownerID,
		ContactID: input.ContactID,
		Type:      input.Type,
		Status:    "pending",
		Reason:    input.Reason,
	}
	if err := db.Create(req).Error; err != nil {
		return nil, err
	}
	return req, nil
}

type DSARInput struct {
	ContactID string  `json:"contact_id" binding:"required"`
	Type      string  `json:"type" binding:"required"`
	Reason    *string `json:"reason"`
}

// ListDSARs returns all DSAR requests for an owner.
func ListDSARs(db *gorm.DB, ownerID string) ([]DSARRequest, error) {
	var out []DSARRequest
	err := db.Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&out).Error
	return out, err
}

// ProcessDSAR moves a DSAR from pending to processing.
func ProcessDSAR(db *gorm.DB, ownerID, id string) (*DSARRequest, error) {
	var req DSARRequest
	if err := db.Where("id = ? AND owner_id = ?", id, ownerID).First(&req).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDSARNotFound
		}
		return nil, err
	}
	req.Status = "processing"
	if err := db.Save(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

// CompleteDSAR marks a DSAR as completed with a response.
func CompleteDSAR(db *gorm.DB, ownerID, id string, response string) (*DSARRequest, error) {
	var req DSARRequest
	if err := db.Where("id = ? AND owner_id = ?", id, ownerID).First(&req).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDSARNotFound
		}
		return nil, err
	}
	now := time.Now()
	req.Status = "completed"
	req.Response = &response
	req.CompletedAt = &now
	if err := db.Save(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

// RejectDSAR marks a DSAR as rejected with a reason.
func RejectDSAR(db *gorm.DB, ownerID, id string, reason string) (*DSARRequest, error) {
	var req DSARRequest
	if err := db.Where("id = ? AND owner_id = ?", id, ownerID).First(&req).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDSARNotFound
		}
		return nil, err
	}
	req.Status = "rejected"
	req.Reason = &reason
	if err := db.Save(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

// ---- Processing Records ----

// LogProcessing creates an audit trail of data processing activities.
func LogProcessing(db *gorm.DB, ownerID string, input *ProcessingInput) error {
	rec := &ProcessingRecord{
		OwnerID:      ownerID,
		ContactID:    input.ContactID,
		Activity:     input.Activity,
		Purpose:      input.Purpose,
		DataCategory: input.DataCategory,
		LegalBasis:   input.LegalBasis,
		Details:      input.Details,
	}
	return db.Create(rec).Error
}

type ProcessingInput struct {
	ContactID    string `json:"contact_id"`
	Activity     string `json:"activity"`
	Purpose      string `json:"purpose"`
	DataCategory string `json:"data_category"`
	LegalBasis   string `json:"legal_basis"`
	Details      string `json:"details"`
}

// ListProcessingRecords returns processing records for an owner (paginated audit).
func ListProcessingRecords(db *gorm.DB, ownerID string, limit int) ([]ProcessingRecord, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var out []ProcessingRecord
	err := db.Where("owner_id = ?", ownerID).Order("created_at DESC").Limit(limit).Find(&out).Error
	return out, err
}

// GetComplianceStats returns counts for the compliance dashboard.
type ComplianceStats struct {
	ActiveConsents   int64 `json:"active_consents"`
	RevokedConsents  int64 `json:"revoked_consents"`
	PendingDSARs     int64 `json:"pending_dsars"`
	CompletedDSARs   int64 `json:"completed_dsars"`
	ProcessingLogs   int64 `json:"processing_logs"`
}

func GetStats(db *gorm.DB, ownerID string) (*ComplianceStats, error) {
	s := &ComplianceStats{}
	db.Model(&ConsentRecord{}).Where("owner_id = ? AND consent_given = ?", ownerID, true).Count(&s.ActiveConsents)
	db.Model(&ConsentRecord{}).Where("owner_id = ? AND consent_given = ?", ownerID, false).Count(&s.RevokedConsents)
	db.Model(&DSARRequest{}).Where("owner_id = ? AND status IN ?", ownerID, []string{"pending", "processing"}).Count(&s.PendingDSARs)
	db.Model(&DSARRequest{}).Where("owner_id = ? AND status = ?", ownerID, "completed").Count(&s.CompletedDSARs)
	db.Model(&ProcessingRecord{}).Where("owner_id = ?", ownerID).Count(&s.ProcessingLogs)
	return s, nil
}
