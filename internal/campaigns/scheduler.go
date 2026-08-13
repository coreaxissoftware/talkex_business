package campaigns

import (
	"log"
	"time"

	"gorm.io/gorm"
)

func StartScheduler(db *gorm.DB) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			launchDueCampaigns(db)
		}
	}()
	log.Println("Campaign scheduler started (30s interval)")
}

func launchDueCampaigns(db *gorm.DB) {
	var due []Campaign
	db.Where("status = ? AND scheduled_at <= ?", StatusScheduled, time.Now()).Find(&due)

	for i := range due {
		if _, err := Launch(db, &due[i]); err != nil {
			log.Printf("scheduler: failed to launch campaign %s: %v", due[i].ID, err)
		} else {
			log.Printf("scheduler: launched campaign %s (%s)", due[i].ID, due[i].Name)
		}
	}
}
