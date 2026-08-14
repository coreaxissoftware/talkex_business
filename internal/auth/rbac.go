package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TeamRoleResolver looks up the role of a user within an owner's team.
// Returns (role, true) if the user is a team member; ("", false) if not.
type TeamRoleResolver func(userID string) (role string, isTeamMember bool, ownerID string)

var teamRoleResolver TeamRoleResolver

func RegisterTeamRoleResolver(f TeamRoleResolver) { teamRoleResolver = f }

// RoleRequired returns middleware that restricts access to users whose
// team role matches one of the allowed roles. The account owner (whose
// user_id matches their own owner_id) is always treated as "admin".
// Non-team users (direct account holders) pass through unblocked —
// they own everything.
func RoleRequired(allowedRoles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		allowed[r] = true
	}

	return func(c *gin.Context) {
		if teamRoleResolver == nil {
			// No team system wired — single-user mode, let everything through.
			c.Next()
			return
		}

		userID := GetUserID(c)
		role, isTeamMember, ownerID := teamRoleResolver(userID)

		if !isTeamMember {
			// Direct account owner — always allowed.
			c.Next()
			return
		}

		// Store the resolved owner_id so downstream handlers can scope
		// data correctly when a team member is acting on behalf of their owner.
		c.Set("resolved_owner_id", ownerID)

		if allowed[role] {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"detail": "Insufficient permissions. Required role: " + joinRoles(allowedRoles),
		})
	}
}

func joinRoles(roles []string) string {
	s := ""
	for i, r := range roles {
		if i > 0 {
			s += " or "
		}
		s += r
	}
	return s
}
