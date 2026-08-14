package organizations

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

var (
	ErrOrgNotFound    = errors.New("organization not found")
	ErrSlugTaken      = errors.New("organization slug already taken")
	ErrMaxSubOrgs     = errors.New("maximum sub-organizations reached")
	ErrMaxUsers       = errors.New("maximum users reached for this organization")
	ErrNotOrgOwner    = errors.New("only the organization owner can perform this action")
)

type CreateInput struct {
	Name     string  `json:"name" binding:"required"`
	Slug     string  `json:"slug" binding:"required"`
	Website  *string `json:"website"`
	ParentID *string `json:"parent_id"`
}

func Create(db *gorm.DB, ownerID string, in *CreateInput) (*Organization, error) {
	slug := strings.ToLower(strings.TrimSpace(in.Slug))

	// Check slug uniqueness
	var count int64
	db.Model(&Organization{}).Where("slug = ?", slug).Count(&count)
	if count > 0 {
		return nil, ErrSlugTaken
	}

	// If creating a sub-org, check parent limits
	if in.ParentID != nil {
		var parent Organization
		if err := db.Where("id = ?", *in.ParentID).First(&parent).Error; err != nil {
			return nil, ErrOrgNotFound
		}
		var subCount int64
		db.Model(&Organization{}).Where("parent_id = ?", parent.ID).Count(&subCount)
		if int(subCount) >= parent.MaxSubOrgs {
			return nil, ErrMaxSubOrgs
		}
	}

	org := &Organization{
		Name:     in.Name,
		Slug:     slug,
		OwnerID:  ownerID,
		ParentID: in.ParentID,
		Website:  in.Website,
	}
	if err := db.Create(org).Error; err != nil {
		return nil, err
	}

	// Auto-add creator as org admin
	member := &OrgMember{
		OrgID:  org.ID,
		UserID: ownerID,
		Role:   "admin",
	}
	db.Create(member)

	return org, nil
}

func GetByID(db *gorm.DB, id string) (*Organization, error) {
	var org Organization
	err := db.Where("id = ?", id).First(&org).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrOrgNotFound
	}
	return &org, err
}

func GetByOwner(db *gorm.DB, ownerID string) ([]Organization, error) {
	var orgs []Organization
	err := db.Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&orgs).Error
	return orgs, err
}

func ListSubOrgs(db *gorm.DB, parentID string) ([]Organization, error) {
	var orgs []Organization
	err := db.Where("parent_id = ?", parentID).Order("name ASC").Find(&orgs).Error
	return orgs, err
}

func ListMembers(db *gorm.DB, orgID string) ([]OrgMember, error) {
	var members []OrgMember
	err := db.Where("org_id = ?", orgID).Find(&members).Error
	return members, err
}

func AddMember(db *gorm.DB, orgID, userID, role string) (*OrgMember, error) {
	// Check org user limit
	var org Organization
	if err := db.Where("id = ?", orgID).First(&org).Error; err != nil {
		return nil, ErrOrgNotFound
	}
	var memberCount int64
	db.Model(&OrgMember{}).Where("org_id = ?", orgID).Count(&memberCount)
	if int(memberCount) >= org.MaxUsers {
		return nil, ErrMaxUsers
	}

	m := &OrgMember{OrgID: orgID, UserID: userID, Role: role}
	if err := db.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

func RemoveMember(db *gorm.DB, orgID, userID string) error {
	return db.Where("org_id = ? AND user_id = ?", orgID, userID).Delete(&OrgMember{}).Error
}

type UpdateInput struct {
	Name       *string  `json:"name"`
	Website    *string  `json:"website"`
	LogoURL    *string  `json:"logo_url"`
	MaxUsers   *int     `json:"max_users"`
	MaxSubOrgs *int     `json:"max_sub_orgs"`
	Tier       *OrgTier `json:"tier"`
}

func Update(db *gorm.DB, org *Organization, in *UpdateInput) (*Organization, error) {
	if in.Name != nil {
		org.Name = *in.Name
	}
	if in.Website != nil {
		org.Website = in.Website
	}
	if in.LogoURL != nil {
		org.LogoURL = in.LogoURL
	}
	if in.MaxUsers != nil {
		org.MaxUsers = *in.MaxUsers
	}
	if in.MaxSubOrgs != nil {
		org.MaxSubOrgs = *in.MaxSubOrgs
	}
	if in.Tier != nil {
		org.Tier = *in.Tier
	}
	return org, db.Save(org).Error
}
