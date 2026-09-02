// Package reseller — cross-tenant margin dashboard for a reseller
// (an Organization with Tier "reseller" and one or more child orgs).
//
// The reseller sees every child tenant as a row: messages sent, what
// the reseller paid the upstream provider (cost), what the reseller
// billed the child (revenue), and the margin = revenue − cost. This
// is the number that matters to the reseller's own P&L, and today it
// requires stitching per-message CostPerMessage + SellPrice on the
// queued_messages table against the organization tree — one endpoint
// so the dashboard is a single round-trip.
package reseller

import (
	"time"

	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/organizations"
)

// TenantRow — one child tenant's roll-up.
type TenantRow struct {
	TenantOwnerID string  `json:"tenant_owner_id"`
	TenantName    string  `json:"tenant_name"`
	Messages      int64   `json:"messages"`
	Cost          float64 `json:"cost"`
	Revenue       float64 `json:"revenue"`
	Margin        float64 `json:"margin"`
	MarginPct     float64 `json:"margin_pct"`
	// PerChannel is a small breakdown so the reseller can see which
	// channel is driving the number.
	PerChannel []ChannelRow `json:"per_channel"`
}

// ChannelRow — same shape as TenantRow but scoped to one channel.
type ChannelRow struct {
	Channel  string  `json:"channel"`
	Messages int64   `json:"messages"`
	Cost     float64 `json:"cost"`
	Revenue  float64 `json:"revenue"`
	Margin   float64 `json:"margin"`
}

// Dashboard is the top-level response.
type Dashboard struct {
	ResellerOwnerID string      `json:"reseller_owner_id"`
	WindowFrom      time.Time   `json:"window_from"`
	WindowTo        time.Time   `json:"window_to"`
	TotalTenants    int         `json:"total_tenants"`
	TotalMessages   int64       `json:"total_messages"`
	TotalCost       float64     `json:"total_cost"`
	TotalRevenue    float64     `json:"total_revenue"`
	TotalMargin     float64     `json:"total_margin"`
	AvgMarginPct    float64     `json:"avg_margin_pct"`
	Tenants         []TenantRow `json:"tenants"`
}

// Build assembles the dashboard for one reseller. The reseller is
// identified by their owner_id — we find the reseller Organization
// they own and then every child organization's owner_id becomes a
// TenantRow. windowDays defaults to 30 when zero or negative.
func Build(db *gorm.DB, resellerOwnerID string, windowDays int) (*Dashboard, error) {
	if windowDays <= 0 {
		windowDays = 30
	}
	windowTo := time.Now()
	windowFrom := windowTo.AddDate(0, 0, -windowDays)

	// 1. Resolve the reseller's organization.
	var reseller organizations.Organization
	err := db.Where("owner_id = ? AND tier = ?", resellerOwnerID, organizations.TierReseller).
		First(&reseller).Error
	if err != nil {
		return nil, err
	}

	// 2. Enumerate child organizations.
	var children []organizations.Organization
	if err := db.Where("parent_id = ?", reseller.ID).Find(&children).Error; err != nil {
		return nil, err
	}

	out := &Dashboard{
		ResellerOwnerID: resellerOwnerID,
		WindowFrom:      windowFrom,
		WindowTo:        windowTo,
		TotalTenants:    len(children),
		Tenants:         make([]TenantRow, 0, len(children)),
	}

	for _, ch := range children {
		row := TenantRow{
			TenantOwnerID: ch.OwnerID,
			TenantName:    ch.Name,
		}
		// Overall roll-up
		db.Raw(
			`SELECT COUNT(*), COALESCE(SUM(cost_per_message),0), COALESCE(SUM(sell_price),0)
			 FROM queued_messages
			 WHERE owner_id = ? AND created_at BETWEEN ? AND ?`,
			ch.OwnerID, windowFrom, windowTo,
		).Row().Scan(&row.Messages, &row.Cost, &row.Revenue)

		row.Margin = row.Revenue - row.Cost
		if row.Revenue > 0 {
			row.MarginPct = (row.Margin / row.Revenue) * 100
		}

		// Per-channel breakdown
		rows, err := db.Raw(
			`SELECT channel, COUNT(*), COALESCE(SUM(cost_per_message),0), COALESCE(SUM(sell_price),0)
			 FROM queued_messages
			 WHERE owner_id = ? AND created_at BETWEEN ? AND ?
			 GROUP BY channel
			 ORDER BY COUNT(*) DESC`,
			ch.OwnerID, windowFrom, windowTo,
		).Rows()
		if err == nil {
			for rows.Next() {
				var cr ChannelRow
				if err := rows.Scan(&cr.Channel, &cr.Messages, &cr.Cost, &cr.Revenue); err == nil {
					cr.Margin = cr.Revenue - cr.Cost
					row.PerChannel = append(row.PerChannel, cr)
				}
			}
			rows.Close()
		}

		out.Tenants = append(out.Tenants, row)
		out.TotalMessages += row.Messages
		out.TotalCost += row.Cost
		out.TotalRevenue += row.Revenue
	}

	out.TotalMargin = out.TotalRevenue - out.TotalCost
	if out.TotalRevenue > 0 {
		out.AvgMarginPct = (out.TotalMargin / out.TotalRevenue) * 100
	}
	return out, nil
}
