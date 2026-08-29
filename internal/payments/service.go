// Package payments provides a Razorpay-compatible payment gateway wrapper
// for wallet top-ups. In dev mode, orders are simulated and can be marked
// as paid via a dev-only endpoint. In production, swap the dev handlers
// for real Razorpay API calls (github.com/razorpay/razorpay-go).
//
// The public API mirrors Razorpay's shape so migration is a drop-in:
//   POST /payments/order        → { order_id, amount, currency, key_id }
//   POST /payments/verify       → verifies checkout signature
//   POST /payments/webhook      → provider posts payment events
//   POST /payments/dev-simulate → dev-only: mark an order paid manually
package payments

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"gorm.io/gorm"
)

var (
	ErrOrderNotFound  = errors.New("payment order not found")
	ErrAlreadyPaid    = errors.New("order already paid")
	ErrOrderExpired   = errors.New("order expired")
	ErrInvalidSignature = errors.New("invalid signature")
)

// PaymentStatus mirrors Razorpay's order lifecycle.
type PaymentStatus string

const (
	StatusCreated PaymentStatus = "created"
	StatusPaid    PaymentStatus = "paid"
	StatusFailed  PaymentStatus = "failed"
)

// Order — an in-flight top-up request. Persisted so we can survive a
// restart between "user clicked pay" and "provider webhook confirms".
type Order struct {
	ID        string        `gorm:"type:varchar(64);primaryKey" json:"id"` // order_xxxx (Razorpay-style)
	OwnerID   string        `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	Amount    float64       `gorm:"not null" json:"amount"`
	Currency  string        `gorm:"type:varchar(8);default:'INR'" json:"currency"`
	Status    PaymentStatus `gorm:"type:varchar(20);not null;default:'created'" json:"status"`
	PaymentID string        `gorm:"type:varchar(64)" json:"payment_id"` // pay_xxxx once paid
	CreatedAt time.Time     `json:"created_at"`
	PaidAt    *time.Time    `json:"paid_at"`
}

// CreditFn is injected by main.go to credit the wallet without a package cycle.
type CreditFn func(ownerID string, amount float64, reference, idempotencyKey string) error

var creditFn CreditFn
var creditMu sync.RWMutex

func RegisterCreditFn(f CreditFn) {
	creditMu.Lock()
	defer creditMu.Unlock()
	creditFn = f
}

// randomID returns a random hex ID with the given prefix.
func randomID(prefix string) string {
	b := make([]byte, 8)
	rand.Read(b)
	return prefix + hex.EncodeToString(b)
}

// CreateOrder starts a new top-up. In production this would call
// Razorpay's Orders API; in dev we just persist a stub.
func CreateOrder(db *gorm.DB, ownerID string, amount float64) (*Order, error) {
	o := &Order{
		ID:       randomID("order_"),
		OwnerID:  ownerID,
		Amount:   amount,
		Currency: "INR",
		Status:   StatusCreated,
	}
	if err := db.Create(o).Error; err != nil {
		return nil, err
	}
	return o, nil
}

// GetOrder returns an order by ID scoped to owner.
func GetOrder(db *gorm.DB, ownerID, orderID string) (*Order, error) {
	var o Order
	if err := db.Where("id = ? AND owner_id = ?", orderID, ownerID).First(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return &o, nil
}

// MarkPaid completes the payment and credits the wallet.
// Used by both the real webhook and the dev-simulate endpoint.
func MarkPaid(db *gorm.DB, orderID, paymentID string) (*Order, error) {
	var o Order
	if err := db.Where("id = ?", orderID).First(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if o.Status == StatusPaid {
		return &o, ErrAlreadyPaid
	}

	now := time.Now()
	o.Status = StatusPaid
	o.PaymentID = paymentID
	o.PaidAt = &now
	if err := db.Save(&o).Error; err != nil {
		return nil, err
	}

	// Credit the wallet — use the paymentID as the idempotency key so
	// a duplicate webhook won't double-credit.
	creditMu.RLock()
	fn := creditFn
	creditMu.RUnlock()
	if fn != nil {
		ref := "Razorpay: " + paymentID
		if err := fn(o.OwnerID, o.Amount, ref, paymentID); err != nil {
			// Payment stays marked paid; wallet credit failure should
			// alert ops — surface via log rather than roll back.
			return &o, err
		}
	}
	return &o, nil
}

// ListOrders returns all orders for an owner, newest first.
func ListOrders(db *gorm.DB, ownerID string) ([]Order, error) {
	var out []Order
	err := db.Where("owner_id = ?", ownerID).Order("created_at DESC").Limit(50).Find(&out).Error
	return out, err
}
