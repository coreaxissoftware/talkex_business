package wallet

import (
	"errors"

	"gorm.io/gorm"
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
func ApplyTransaction(db *gorm.DB, w *Wallet, txnType TransactionType, amount float64, idempotencyKey string, reference *string) (*WalletTransaction, error) {
	// Check idempotency — return existing if already applied.
	var existing WalletTransaction
	err := db.Where("wallet_id = ? AND idempotency_key = ?", w.ID, idempotencyKey).First(&existing).Error
	if err == nil {
		return &existing, nil
	}

	var delta float64
	if txnType == Credit {
		delta = amount
	} else {
		delta = -amount
	}

	newBalance := w.Balance + delta
	if newBalance < 0 {
		return nil, ErrInsufficientBalance
	}

	txn := &WalletTransaction{
		WalletID:       w.ID,
		Type:           txnType,
		Amount:         amount,
		BalanceAfter:   newBalance,
		Reference:      reference,
		IdempotencyKey: idempotencyKey,
	}

	// Use a DB transaction to atomically update balance + insert txn row.
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(w).Update("balance", newBalance).Error; err != nil {
			return err
		}
		return tx.Create(txn).Error
	})
	if err != nil {
		return nil, err
	}

	w.Balance = newBalance
	return txn, nil
}
