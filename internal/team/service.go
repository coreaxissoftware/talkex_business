package team

import (
	"errors"
	"time"

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

// AgentActivity aggregates per-agent workload for the Team Activity
// dashboard. Counts are over the last 30 days.
type AgentActivity struct {
	UserID        string  `json:"user_id"`
	Name          string  `json:"name"`
	Email         string  `json:"email"`
	Role          string  `json:"role"`
	OpenAssigned  int64   `json:"open_assigned"`
	MessagesSent  int64   `json:"messages_sent_30d"`
	AvgCSAT       float64 `json:"avg_csat_30d"`
	CSATCount     int64   `json:"csat_count_30d"`
}

// Activity returns each team member's workload plus the account
// owner as an implicit agent (id = ownerID). Raw SQL joins are
// avoided to keep the func portable across sqlite + postgres.
func Activity(db *gorm.DB, ownerID string) ([]AgentActivity, error) {
	members, err := List(db, ownerID)
	if err != nil {
		return nil, err
	}
	// The account owner is always an agent even when not in `members`.
	rows := []AgentActivity{{UserID: ownerID, Name: "Account owner", Role: "owner"}}
	for _, m := range members {
		if m.Status != "active" {
			continue
		}
		rows = append(rows, AgentActivity{
			UserID: m.UserID,
			Name:   m.Name,
			Email:  m.Email,
			Role:   string(m.Role),
		})
	}

	// Populate counts per row. Two queries per agent is fine at
	// team sizes (<100 members typical). Time cutoff is passed as
	// a parameter so this works on both sqlite and postgres.
	since30 := time.Now().AddDate(0, 0, -30)
	for i := range rows {
		uid := rows[i].UserID
		if uid == "" {
			continue
		}
		db.Raw(
			`SELECT COUNT(*) FROM conversations WHERE owner_id = ? AND assigned_to = ?`,
			ownerID, uid,
		).Scan(&rows[i].OpenAssigned)

		db.Raw(
			`SELECT COUNT(*) FROM messages m
			 JOIN conversations c ON c.id = m.conversation_id
			 WHERE c.owner_id = ? AND c.assigned_to = ?
			   AND m.direction = 'outbound' AND m.created_at >= ?`,
			ownerID, uid, since30,
		).Scan(&rows[i].MessagesSent)

		var avg float64
		var cnt int64
		db.Raw(
			`SELECT COUNT(*), COALESCE(AVG(score), 0) FROM ratings
			 WHERE owner_id = ? AND agent_user_id = ? AND created_at >= ?`,
			ownerID, uid, since30,
		).Row().Scan(&cnt, &avg)
		rows[i].CSATCount = cnt
		rows[i].AvgCSAT = avg
	}
	return rows, nil
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

// GetByUserID finds a team membership by the linked user_id (after accepting
// an invite). Returns nil if the user is not on any team.
func GetByUserID(db *gorm.DB, userID string) (*Member, error) {
	var m Member
	err := db.Where("user_id = ? AND status = ?", userID, StatusActive).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &m, err
}

func Remove(db *gorm.DB, m *Member) error {
	return db.Delete(m).Error
}
