package settings

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

func Get(db *gorm.DB, ownerID string) (*UserSettings, *PrefsData, error) {
	var s UserSettings
	err := db.Where("owner_id = ?", ownerID).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		defaults := defaultPrefs()
		return &UserSettings{OwnerID: ownerID}, defaults, nil
	}
	if err != nil {
		return nil, nil, err
	}
	var prefs PrefsData
	json.Unmarshal(s.Prefs, &prefs)
	return &s, &prefs, nil
}

func Save(db *gorm.DB, ownerID string, prefs *PrefsData) (*UserSettings, error) {
	var s UserSettings
	err := db.Where("owner_id = ?", ownerID).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		s = UserSettings{OwnerID: ownerID}
	} else if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(prefs)
	s.Prefs = data

	if s.ID == "" {
		return &s, db.Create(&s).Error
	}
	return &s, db.Save(&s).Error
}

func defaultPrefs() *PrefsData {
	return &PrefsData{
		NotifCampaigns: true,
		NotifMessages:  true,
		NotifSystem:    true,
		EmailDigest:    false,
		Timezone:       "Asia/Kolkata",
		Language:       "en",
	}
}
