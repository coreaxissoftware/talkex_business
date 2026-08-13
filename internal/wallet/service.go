package wallet

import (
	"errors"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrInsufficientBalance = errors.New("wallet balance cannot go negative")
)

func GetOrCreateWallet(db *gorm.DB, userID string) (*Wallet, error) {
	var w Wallet
	err := db.Where("user_id = ?", userID).First(&w).Error
	if err == nil {
		return &w, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	w = Wallet{UserID: userID, Currency: "INR"}
	if err := db.Create(&w).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func ListTransactions(db *gorm.DB, walletID string) ([]WalletTransaction, error) {
	var txns []WalletTransaction
	err := db.Where("wallet_id = ?", walletID).Order("created_at DESC").Find(&txns).Error
	return txns, err
}

// ApplyTransaction credits/debits a wallet, guarded by idempotency_key so
// a retried call returns the original result instead of applying twice.
//
// Concurrency-safe: the whole read-check-write path runs inside a DB
// transaction with a row-level lock on the wallet (Postgres SELECT FOR
// UPDATE; SQLite BEGIN IMMEDIATE gives an equivalent single-writer
// guarantee via glebarez/sqlite). A racing duplicate-key insert is
// converted back into a return of the existing row.
func ApplyTransaction(db *gorm.DB, w *Wallet, txnType TransactionType, amount float64, idempotencyKey string, reference *string) (*WalletTransaction, error) {
	if amount < 0 {
		// Callers pass positive amounts; direction lives in txnType.
		amount = -amount
	}

	var out *WalletTransaction
	err := db.Transaction(func(tx *gorm.DB) error {
		// Idempotency: same key against the same wallet returns the prior row.
		var existing WalletTransaction
		if err := tx.Where("wallet_id = ? AND idempotency_key = ?", w.ID, idempotencyKey).
			First(&existing).Error; err == nil {
			out = &existing
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// Re-read the wallet under a row lock so the balance we compute
		// against is authoritative for the life of this transaction.
		var locked Wallet
		q := tx.Where("id = ?", w.ID)
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
		if err := q.First(&locked).Error; err != nil {
			return err
		}

		delta := amount
		if txnType == Debit {
			delta = -amount
		}
		newBalance := locked.Balance + delta
		if newBalance < 0 {
			return ErrInsufficientBalance
		}

		txn := &WalletTransaction{
			WalletID:       locked.ID,
			Type:           txnType,
			Amount:         amount,
			BalanceAfter:   newBalance,
			Reference:      reference,
			IdempotencyKey: idempotencyKey,
		}

		if err := tx.Model(&locked).Update("balance", newBalance).Error; err != nil {
			return err
		}
		if err := tx.Create(txn).Error; err != nil {
			// A concurrent retry with the same key raced us — return that row.
			if isUniqueViolation(err) {
				var existing WalletTransaction
				if fetchErr := tx.Where("wallet_id = ? AND idempotency_key = ?", locked.ID, idempotencyKey).
					First(&existing).Error; fetchErr == nil {
					out = &existing
					return nil
				}
			}
			return err
		}

		// Mirror the new balance onto the caller's Wallet pointer so subsequent
		// reads of `w.Balance` reflect the applied change.
		w.Balance = newBalance
		out = txn
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// isUniqueViolation is best-effort string matching — the underlying error
// text differs between Postgres (`SQLSTATE 23505`, `duplicate key`) and
// SQLite (`UNIQUE constraint failed`). Either counts as "someone else won
// the race".
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "duplicate key value") ||
		strings.Contains(msg, "23505")
}
