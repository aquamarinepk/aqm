package auth

import (
	"time"

	"github.com/google/uuid"
)

// Grant represents a role assignment to a user.
// NOTE: With username-based grants, AuthZ no longer requires UUID synchronization
// with AuthN. Each service maintains its own internal identifiers.
// The username serves as the natural key for cross-service correlation.
type Grant struct {
	ID         uuid.UUID `json:"id" db:"id" bson:"_id"`
	Username   string    `json:"username" db:"username" bson:"username"`
	RoleID     uuid.UUID `json:"role_id" db:"role_id" bson:"role_id"`
	AssignedAt time.Time `json:"assigned_at" db:"assigned_at" bson:"assigned_at"`
	AssignedBy string    `json:"assigned_by" db:"assigned_by" bson:"assigned_by"`
}

func NewGrant(username string, roleID uuid.UUID, assignedBy string) *Grant {
	return &Grant{
		ID:         uuid.New(),
		Username:   username,
		RoleID:     roleID,
		AssignedAt: time.Now(),
		AssignedBy: assignedBy,
	}
}

func (g *Grant) Validate() error {
	if g.Username == "" {
		return ErrUserNotFound
	}
	if g.RoleID == uuid.Nil {
		return ErrRoleNotFound
	}
	if g.AssignedBy == "" {
		return ErrPermissionDenied
	}
	return nil
}
