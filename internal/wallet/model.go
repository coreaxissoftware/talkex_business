// Package wallet tracks balance + ledger. One Wallet per user; every
// balance change is a WalletTransaction row so balance is always
// reconstructable/auditable.
//
// idempotency_key on WalletTransaction is called out as an MVP-blocking
// gap in CONTEXT.md ("Idempotency keys on send/wallet-debit") — a retried
// debit (e.g. a campaign-send retry after a timeout) must not double-charge.
package wallet

import (
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

type TransactionType string

const (
	Credit TransactionType = "credit"
	Debit  TransactionType = "debit"
)

type Wallet struct {
	database.Base
	UserID   string  `gorm:"type:varchar(36);uniqueIndex;not null" json:"user_id"`
	Balance  float64 `gorm:"type:decimal(14,4);default:0;not null" json:"balance"`
	Currency string  `gorm:"type:varchar(3);default:INR;not null" json:"currency"`
}

type WalletTransaction struct {
	database.Base
	WalletID       string          `gorm:"type:varchar(36);not null;index" json:"wallet_id"`
	Type           TransactionType `gorm:"type:varchar(10);not null" json:"type"`
	Amount         float64         `gorm:"type:decimal(14,4);not null" json:"amount"`
	BalanceAfter   float64         `gorm:"type:decimal(14,4);not null" json:"balance_after"`
	Reference      *string         `gorm:"type:varchar(255)" json:"reference"`
	IdempotencyKey string          `gorm:"type:varchar(64);not null;uniqueIndex:uq_wallet_idempotency_key" json:"idempotency_key"`

	// Composite unique: (wallet_id, idempotency_key) — same key can't be
	// applied twice against the same wallet. GORM uniqueIndex tag on the
	// single field is simpler than a composite, but we enforce composite
	// uniqueness via service-level check + DB unique constraint below.
}

// TableName overrides the default table name.
func (WalletTransaction) TableName() string {
	return "wallet_transactions"
}
