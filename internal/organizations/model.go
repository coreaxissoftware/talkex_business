// Package organizations implements multi-tenancy and reseller hierarchy.
// An Organization owns users, and can have a parent org (reseller model).
// Users belong to an org via OrgID; the org owner can manage sub-accounts.
package organizations

import (
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

type OrgTier string

const (
	TierFree       OrgTier = "free"
	TierStarter    OrgTier = "starter"
	TierBusiness   OrgTier = "business"
	TierEnterprise OrgTier = "enterprise"
	TierReseller   OrgTier = "reseller"
)

type Organization struct {
	database.Base
	Name        string  `gorm:"type:varchar(255);not null" json:"name"`
	Slug        string  `gorm:"type:varchar(100);uniqueIndex;not null" json:"slug"`
	OwnerID     string  `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	ParentID    *string `gorm:"type:varchar(36);index" json:"parent_id"`
	Tier        OrgTier `gorm:"type:varchar(20);default:'free';not null" json:"tier"`
	LogoURL     *string `gorm:"type:varchar(500)" json:"logo_url"`
	Website     *string `gorm:"type:varchar(500)" json:"website"`
	MaxUsers    int     `gorm:"default:5;not null" json:"max_users"`
	MaxSubOrgs  int     `gorm:"default:0;not null" json:"max_sub_orgs"`
	IsActive    bool    `gorm:"default:true;not null" json:"is_active"`
}

// OrgMember links a user to an organization with a role.
type OrgMember struct {
	database.Base
	OrgID  string `gorm:"type:varchar(36);index;not null;uniqueIndex:uq_org_user" json:"org_id"`
	UserID string `gorm:"type:varchar(36);index;not null;uniqueIndex:uq_org_user" json:"user_id"`
	Role   string `gorm:"type:varchar(20);default:'member';not null" json:"role"`
}
