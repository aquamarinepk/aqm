package auth

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewGrant(t *testing.T) {
	userID := uuid.New()
	roleID := uuid.New()
	assignedBy := "admin"

	grant := NewGrant(userID, roleID, assignedBy)

	if grant == nil {
		t.Fatal("NewGrant() returned nil")
	}
	if grant.ID == uuid.Nil {
		t.Error("NewGrant() did not generate ID")
	}
	if grant.UserID != userID {
		t.Errorf("NewGrant() userID = %v, want %v", grant.UserID, userID)
	}
	if grant.RoleID != roleID {
		t.Errorf("NewGrant() roleID = %v, want %v", grant.RoleID, roleID)
	}
	if grant.AssignedBy != assignedBy {
		t.Errorf("NewGrant() assignedBy = %v, want %v", grant.AssignedBy, assignedBy)
	}
	if grant.AssignedAt.IsZero() {
		t.Error("NewGrant() did not set AssignedAt")
	}
}

func TestGrantValidate(t *testing.T) {
	userID := uuid.New()
	roleID := uuid.New()

	tests := []struct {
		name    string
		grant   *Grant
		wantErr bool
	}{
		{
			"valid grant",
			&Grant{UserID: userID, RoleID: roleID, AssignedBy: "admin"},
			false,
		},
		{
			"missing user ID",
			&Grant{UserID: uuid.Nil, RoleID: roleID, AssignedBy: "admin"},
			true,
		},
		{
			"missing role ID",
			&Grant{UserID: userID, RoleID: uuid.Nil, AssignedBy: "admin"},
			true,
		},
		{
			"missing assigned by",
			&Grant{UserID: userID, RoleID: roleID, AssignedBy: ""},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.grant.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
