package tags

import (
	"encoding/json"
	"strings"

	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/contacts"
)

type TagCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func ListAll(db *gorm.DB, ownerID string) ([]TagCount, error) {
	var all []contacts.Contact
	if err := db.Where("owner_id = ?", ownerID).Select("tags").Find(&all).Error; err != nil {
		return nil, err
	}

	counts := map[string]int{}
	for _, c := range all {
		var tags []string
		_ = json.Unmarshal(c.Tags, &tags)
		for _, t := range tags {
			counts[strings.TrimSpace(t)]++
		}
	}

	out := make([]TagCount, 0, len(counts))
	for name, count := range counts {
		if name == "" {
			continue
		}
		out = append(out, TagCount{Name: name, Count: count})
	}
	return out, nil
}

func Rename(db *gorm.DB, ownerID, oldName, newName string) (int, error) {
	var all []contacts.Contact
	if err := db.Where("owner_id = ?", ownerID).Find(&all).Error; err != nil {
		return 0, err
	}

	updated := 0
	for _, c := range all {
		var tags []string
		_ = json.Unmarshal(c.Tags, &tags)
		found := false
		for i, t := range tags {
			if t == oldName {
				tags[i] = newName
				found = true
			}
		}
		if found {
			b, _ := json.Marshal(tags)
			if err := db.Model(&c).Update("tags", b).Error; err != nil {
				return updated, err
			}
			updated++
		}
	}
	return updated, nil
}

func Delete(db *gorm.DB, ownerID, tagName string) (int, error) {
	var all []contacts.Contact
	if err := db.Where("owner_id = ?", ownerID).Find(&all).Error; err != nil {
		return 0, err
	}

	updated := 0
	for _, c := range all {
		var tags []string
		_ = json.Unmarshal(c.Tags, &tags)
		filtered := make([]string, 0, len(tags))
		for _, t := range tags {
			if t != tagName {
				filtered = append(filtered, t)
			}
		}
		if len(filtered) != len(tags) {
			b, _ := json.Marshal(filtered)
			if err := db.Model(&c).Update("tags", b).Error; err != nil {
				return updated, err
			}
			updated++
		}
	}
	return updated, nil
}

func BulkApply(db *gorm.DB, ownerID, tagName string, contactIDs []string) (int, error) {
	var targets []contacts.Contact
	if err := db.Where("owner_id = ? AND id IN ?", ownerID, contactIDs).Find(&targets).Error; err != nil {
		return 0, err
	}

	applied := 0
	for _, c := range targets {
		var tags []string
		_ = json.Unmarshal(c.Tags, &tags)
		exists := false
		for _, t := range tags {
			if t == tagName {
				exists = true
				break
			}
		}
		if !exists {
			tags = append(tags, tagName)
			b, _ := json.Marshal(tags)
			if err := db.Model(&c).Update("tags", b).Error; err != nil {
				return applied, err
			}
			applied++
		}
	}
	return applied, nil
}
