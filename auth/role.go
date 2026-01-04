package auth

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID  `json:"id" db:"id" bson:"_id"`
	Name        string     `json:"name" db:"name" bson:"name"`
	Description string     `json:"description" db:"description" bson:"description"`
	Permissions []string   `json:"permissions" db:"permissions" bson:"permissions"`
	Status      RoleStatus `json:"status" db:"status" bson:"status"`

	CreatedAt time.Time `json:"created_at" db:"created_at" bson:"created_at"`
	CreatedBy string    `json:"created_by" db:"created_by" bson:"created_by"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" bson:"updated_at"`
	UpdatedBy string    `json:"updated_by" db:"updated_by" bson:"updated_by"`
}

func NewRole() *Role {
	return &Role{
		Status:      RoleStatusActive,
		Permissions: []string{},
	}
}

func (r *Role) EnsureID() {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
}

func (r *Role) BeforeCreate() {
	r.EnsureID()
	now := time.Now()
	r.CreatedAt = now
	r.UpdatedAt = now
	r.Name = NormalizeRoleName(r.Name)
	r.Description = NormalizeDisplayName(r.Description)
	if r.Permissions == nil {
		r.Permissions = []string{}
	}
}

func (r *Role) BeforeUpdate() {
	r.UpdatedAt = time.Now()
	r.Name = NormalizeRoleName(r.Name)
	r.Description = NormalizeDisplayName(r.Description)
	if r.Permissions == nil {
		r.Permissions = []string{}
	}
}

func (r *Role) HasPermission(permission string) bool {
	return HasPermission(r.Permissions, permission)
}

func (r *Role) AddPermission(permission string) {
	if !r.HasPermission(permission) {
		r.Permissions = append(r.Permissions, permission)
	}
}

func (r *Role) RemovePermission(permission string) {
	filtered := make([]string, 0, len(r.Permissions))
	for _, p := range r.Permissions {
		if p != permission {
			filtered = append(filtered, p)
		}
	}
	r.Permissions = filtered
}

func (r *Role) Validate() error {
	if err := ValidateRoleName(r.Name); err != nil {
		return err
	}
	if !r.Status.IsValid() {
		return ErrInvalidRoleName
	}
	return nil
}
