package auth

import (
	"time"

	"github.com/google/uuid"
)

type Grant struct {
	ID         uuid.UUID `json:"id" db:"id" bson:"_id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id" bson:"user_id"`
	RoleID     uuid.UUID `json:"role_id" db:"role_id" bson:"role_id"`
	AssignedAt time.Time `json:"assigned_at" db:"assigned_at" bson:"assigned_at"`
	AssignedBy string    `json:"assigned_by" db:"assigned_by" bson:"assigned_by"`
}

func NewGrant(userID, roleID uuid.UUID, assignedBy string) *Grant {
	return &Grant{
		ID:         uuid.New(),
		UserID:     userID,
		RoleID:     roleID,
		AssignedAt: time.Now(),
		AssignedBy: assignedBy,
	}
}

func (g *Grant) Validate() error {
	if g.UserID == uuid.Nil {
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
