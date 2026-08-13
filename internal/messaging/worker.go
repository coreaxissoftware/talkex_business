package messaging

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// StartWorker runs a background goroutine that processes the message queue
// every interval. It picks messages ordered by priority and creation time.
func StartWorker(db *gorm.DB, interval time.Duration, batchSize int) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		sweepTick := 0
		for range ticker.C {
			n := ProcessQueue(db, batchSize)
			if n > 0 {
				log.Printf("messaging: processed %d messages", n)
			}
			sweepTick++
			if sweepTick%12 == 0 {
				moved := SweepToDLQ(db)
				if moved > 0 {
					log.Printf("messaging: moved %d failed messages to DLQ", moved)
				}
			}
		}
	}()
}
