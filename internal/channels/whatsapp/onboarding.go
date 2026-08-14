package whatsapp

import (
	"errors"

	"github.com/coreaxissoftware/talkex_business/internal/database"
	"gorm.io/gorm"
)

// Onboarding tracks the multi-step WABA (WhatsApp Business Account) setup
// process: business verification → phone number registration → display
// name review → go live.
//
// Steps: business_info → verification → phone_registration → display_name → completed

const (
	StepBusinessInfo      = "business_info"
	StepVerification      = "verification"
	StepPhoneRegistration = "phone_registration"
	StepDisplayName       = "display_name"
	StepCompleted         = "completed"
)

type Onboarding struct {
	database.Base
	OwnerID             string  `gorm:"type:varchar(36);uniqueIndex;not null" json:"owner_id"`
	CurrentStep         string  `gorm:"type:varchar(30);default:'business_info';not null" json:"current_step"`
	BusinessName        string  `gorm:"type:varchar(255)" json:"business_name"`
	BusinessWebsite     string  `gorm:"type:varchar(500)" json:"business_website"`
	BusinessCategory    string  `gorm:"type:varchar(100)" json:"business_category"`
	BusinessAddress     string  `gorm:"type:varchar(500)" json:"business_address"`
	FBBusinessManagerID *string `gorm:"type:varchar(100)" json:"fb_business_manager_id"`
	VerificationStatus  string  `gorm:"type:varchar(20);default:'pending'" json:"verification_status"`
	PhoneNumber         *string `gorm:"type:varchar(20)" json:"phone_number"`
	PhoneVerified       bool    `gorm:"default:false;not null" json:"phone_verified"`
	DisplayName         *string `gorm:"type:varchar(255)" json:"display_name"`
	DisplayNameStatus   string  `gorm:"type:varchar(20);default:'pending'" json:"display_name_status"`
	WABAId              *string `gorm:"type:varchar(100)" json:"waba_id"`
	PhoneNumberID       *string `gorm:"type:varchar(100)" json:"phone_number_id"`
}

var stepOrder = []string{
	StepBusinessInfo,
	StepVerification,
	StepPhoneRegistration,
	StepDisplayName,
	StepCompleted,
}

func GetOnboarding(db *gorm.DB, ownerID string) (*Onboarding, error) {
	var o Onboarding
	err := db.Where("owner_id = ?", ownerID).First(&o).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &o, err
}

func StartOnboarding(db *gorm.DB, ownerID string) (*Onboarding, error) {
	// Check if one already exists
	existing, err := GetOnboarding(db, ownerID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}
	o := &Onboarding{
		OwnerID:     ownerID,
		CurrentStep: StepBusinessInfo,
	}
	return o, db.Create(o).Error
}

type BusinessInfoInput struct {
	BusinessName     string `json:"business_name" binding:"required"`
	BusinessWebsite  string `json:"business_website"`
	BusinessCategory string `json:"business_category"`
	BusinessAddress  string `json:"business_address"`
}

func SaveBusinessInfo(db *gorm.DB, o *Onboarding, in *BusinessInfoInput) error {
	o.BusinessName = in.BusinessName
	o.BusinessWebsite = in.BusinessWebsite
	o.BusinessCategory = in.BusinessCategory
	o.BusinessAddress = in.BusinessAddress
	o.CurrentStep = StepVerification
	return db.Save(o).Error
}

type VerificationInput struct {
	FBBusinessManagerID string `json:"fb_business_manager_id" binding:"required"`
}

func SaveVerification(db *gorm.DB, o *Onboarding, in *VerificationInput) error {
	o.FBBusinessManagerID = &in.FBBusinessManagerID
	o.VerificationStatus = "submitted"
	o.CurrentStep = StepPhoneRegistration
	return db.Save(o).Error
}

type PhoneRegistrationInput struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

func SavePhoneRegistration(db *gorm.DB, o *Onboarding, in *PhoneRegistrationInput) error {
	o.PhoneNumber = &in.PhoneNumber
	o.PhoneVerified = true // In production, OTP verification would happen here
	o.CurrentStep = StepDisplayName
	return db.Save(o).Error
}

type DisplayNameInput struct {
	DisplayName string `json:"display_name" binding:"required"`
}

func SaveDisplayName(db *gorm.DB, o *Onboarding, in *DisplayNameInput) error {
	o.DisplayName = &in.DisplayName
	o.DisplayNameStatus = "submitted"
	o.CurrentStep = StepCompleted
	return db.Save(o).Error
}
