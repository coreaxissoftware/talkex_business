package messaging

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// WalletChecker is injected from main.go to avoid importing wallet directly.
type WalletChecker func(ownerID string) (balance float64, minBalance float64, autoPauseEnabled bool)

// PauseCallback is called when low-wallet auto-pause triggers.
type PauseCallback func(ownerID string, balance float64)

var (
	walletChecker WalletChecker
	pauseCallback PauseCallback
)

func RegisterWalletChecker(f WalletChecker) { walletChecker = f }
func RegisterPauseCallback(f PauseCallback) { pauseCallback = f }

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
			// Check wallet balances every 6th tick (~30s) for auto-pause
			if sweepTick%6 == 0 {
				checkWalletAutoPause(db)
			}
		}
	}()
}

// checkWalletAutoPause finds owners with queued messages and pauses their
// campaigns if wallet balance is below the configured minimum.
func checkWalletAutoPause(db *gorm.DB) {
	if walletChecker == nil {
		return
	}

	// Find distinct owners with queued messages
	var ownerIDs []string
	db.Model(&QueuedMessage{}).
		Where("status = ?", "queued").
		Distinct("owner_id").
		Pluck("owner_id", &ownerIDs)

	for _, ownerID := range ownerIDs {
		balance, minBalance, enabled := walletChecker(ownerID)
		if !enabled || minBalance <= 0 {
			continue
		}
		if balance < minBalance {
			log.Printf("messaging: auto-pause triggered for owner %s (balance=%.2f, min=%.2f)", ownerID, balance, minBalance)
			if pauseCallback != nil {
				pauseCallback(ownerID, balance)
			}
		}
	}
}
