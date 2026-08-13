package billing

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var ErrUnknownPlan = errors.New("unknown plan")

// GetOrCreate returns the owner's Subscription — creating it on the
// starter plan for a brand new owner (mirrors the wallet.GetOrCreateWallet
// pattern used elsewhere).
func GetOrCreate(db *gorm.DB, ownerID string) (*Subscription, error) {
	var s Subscription
	err := db.Where("owner_id = ?", ownerID).First(&s).Error
	if err == nil {
		return &s, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	s = Subscription{
		OwnerID:     ownerID,
		Plan:        PlanStarter,
		PeriodStart: time.Now(),
		Status:      "active",
	}
	if err := db.Create(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// ChangePlan is the self-serve upgrade / downgrade path. In this MVP it
// simply swaps the plan and immediately issues a "paid" invoice for the
// new plan's fee — a real payment provider swap is a Phase 3 gap.
func ChangePlan(db *gorm.DB, ownerID string, planID PlanID) (*Subscription, *Invoice, error) {
	plan := PlanByID(planID)
	if plan == nil {
		return nil, nil, ErrUnknownPlan
	}
	sub, err := GetOrCreate(db, ownerID)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now()
	invoice := &Invoice{
		OwnerID:      ownerID,
		Plan:         planID,
		PeriodStart:  now,
		PeriodEnd:    now.AddDate(0, 1, 0),
		MessagesUsed: 0,
		AmountINR:    float64(plan.PriceINRPerMonth),
		Status:       "paid",
		PaidAt:       &now,
	}
	if err := db.Create(invoice).Error; err != nil {
		return nil, nil, err
	}

	sub.Plan = planID
	sub.PeriodStart = now
	sub.MessagesUsed = 0
	sub.Status = "active"
	sub.CurrentInvoiceID = &invoice.ID
	if err := db.Save(sub).Error; err != nil {
		return nil, nil, err
	}
	return sub, invoice, nil
}

// ListInvoices returns paid + pending invoices newest-first.
func ListInvoices(db *gorm.DB, ownerID string) ([]Invoice, error) {
	var out []Invoice
	err := db.Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&out).Error
	return out, err
}
