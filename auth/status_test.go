package auth

import "testing"

func TestUserStatusString(t *testing.T) {
	tests := []struct {
		name   string
		status UserStatus
		want   string
	}{
		{"active", UserStatusActive, "active"},
		{"suspended", UserStatusSuspended, "suspended"},
		{"pending", UserStatusPending, "pending"},
		{"deleted", UserStatusDeleted, "deleted"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("UserStatus.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserStatusIsValid(t *testing.T) {
	tests := []struct {
		name   string
		status UserStatus
		want   bool
	}{
		{"active is valid", UserStatusActive, true},
		{"suspended is valid", UserStatusSuspended, true},
		{"pending is valid", UserStatusPending, true},
		{"deleted is valid", UserStatusDeleted, true},
		{"empty is invalid", UserStatus(""), false},
		{"unknown is invalid", UserStatus("unknown"), false},
		{"random is invalid", UserStatus("random"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("UserStatus.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleStatusString(t *testing.T) {
	tests := []struct {
		name   string
		status RoleStatus
		want   string
	}{
		{"active", RoleStatusActive, "active"},
		{"inactive", RoleStatusInactive, "inactive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("RoleStatus.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleStatusIsValid(t *testing.T) {
	tests := []struct {
		name   string
		status RoleStatus
		want   bool
	}{
		{"active is valid", RoleStatusActive, true},
		{"inactive is valid", RoleStatusInactive, true},
		{"empty is invalid", RoleStatus(""), false},
		{"unknown is invalid", RoleStatus("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("RoleStatus.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
