package team

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrMemberNotFound  = errors.New("team member not found")
	ErrAlreadyInvited  = errors.New("this email is already invited")
	ErrInvalidRole     = errors.New("invalid role; must be admin, agent, or viewer")
)

func List(db *gorm.DB, ownerID string) ([]Member, error) {
	var out []Member
	err := db.Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&out).Error
	return out, err
}

func GetByID(db *gorm.DB, ownerID, id string) (*Member, error) {
	var m Member
	err := db.Where("id = ? AND owner_id = ?", id, ownerID).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMemberNotFound
	}
	return &m, err
}

type InviteInput struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name"`
	Role  string `json:"role" binding:"required"`
}

func Invite(db *gorm.DB, ownerID string, in *InviteInput) (*Member, error) {
	if in.Role != RoleAdmin && in.Role != RoleAgent && in.Role != RoleViewer {
		return nil, ErrInvalidRole
	}

	var count int64
	db.Model(&Member{}).Where("owner_id = ? AND email = ?", ownerID, in.Email).Count(&count)
	if count > 0 {
		return nil, ErrAlreadyInvited
	}

	m := &Member{
		OwnerID: ownerID,
		Email:   in.Email,
		Name:    in.Name,
		Role:    in.Role,
		Status:  StatusPending,
	}
	if err := db.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

type UpdateRoleInput struct {
	Role string `json:"role" binding:"required"`
}

func UpdateRole(db *gorm.DB, m *Member, role string) (*Member, error) {
	if role != RoleAdmin && role != RoleAgent && role != RoleViewer {
		return nil, ErrInvalidRole
	}
	m.Role = role
	return m, db.Save(m).Error
}

func Remove(db *gorm.DB, m *Member) error {
	return db.Delete(m).Error
}
