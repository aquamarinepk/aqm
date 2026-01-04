package auth

import "testing"

func TestPermissionString(t *testing.T) {
	tests := []struct {
		name string
		perm Permission
		want string
	}{
		{"simple permission", Permission("users:read"), "users:read"},
		{"wildcard", Permission("*"), "*"},
		{"namespace wildcard", Permission("users:*"), "users:*"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.perm.String(); got != tt.want {
				t.Errorf("Permission.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermissionMatches(t *testing.T) {
	tests := []struct {
		name     string
		perm     Permission
		required Permission
		want     bool
	}{
		{"exact match", Permission("users:read"), Permission("users:read"), true},
		{"wildcard matches all", Permission("*"), Permission("users:read"), true},
		{"wildcard matches anything", Permission("*"), Permission("orders:write"), true},
		{"namespace wildcard matches action", Permission("users:*"), Permission("users:read"), true},
		{"namespace wildcard matches different action", Permission("users:*"), Permission("users:write"), true},
		{"action wildcard matches namespace", Permission("*:read"), Permission("users:read"), true},
		{"action wildcard matches different namespace", Permission("*:read"), Permission("orders:read"), true},
		{"no match different namespace", Permission("users:read"), Permission("orders:read"), false},
		{"no match different action", Permission("users:read"), Permission("users:write"), false},
		{"no match different length", Permission("users:read"), Permission("users"), false},
		{"no match different length reverse", Permission("users"), Permission("users:read"), false},
		{"double wildcard matches complex", Permission("*:*"), Permission("users:read"), true},
		{"partial wildcard no match", Permission("user:*"), Permission("users:read"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.perm.Matches(tt.required); got != tt.want {
				t.Errorf("Permission.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasPermission(t *testing.T) {
	tests := []struct {
		name        string
		permissions []string
		required    string
		want        bool
	}{
		{"has exact permission", []string{"users:read", "users:write"}, "users:read", true},
		{"has wildcard", []string{"*"}, "users:read", true},
		{"has namespace wildcard", []string{"users:*"}, "users:read", true},
		{"does not have permission", []string{"users:read"}, "users:write", false},
		{"empty permissions", []string{}, "users:read", false},
		{"no match in list", []string{"orders:read", "products:read"}, "users:read", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasPermission(tt.permissions, tt.required); got != tt.want {
				t.Errorf("HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasAnyPermission(t *testing.T) {
	tests := []struct {
		name        string
		permissions []string
		required    []string
		want        bool
	}{
		{"has one of required", []string{"users:read"}, []string{"users:read", "users:write"}, true},
		{"has all required", []string{"users:read", "users:write"}, []string{"users:read", "users:write"}, true},
		{"has none", []string{"orders:read"}, []string{"users:read", "users:write"}, false},
		{"empty required", []string{"users:read"}, []string{}, false},
		{"empty permissions", []string{}, []string{"users:read"}, false},
		{"wildcard has any", []string{"*"}, []string{"users:read", "orders:write"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasAnyPermission(tt.permissions, tt.required); got != tt.want {
				t.Errorf("HasAnyPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasAllPermissions(t *testing.T) {
	tests := []struct {
		name        string
		permissions []string
		required    []string
		want        bool
	}{
		{"has all required", []string{"users:read", "users:write"}, []string{"users:read", "users:write"}, true},
		{"has more than required", []string{"users:read", "users:write", "orders:read"}, []string{"users:read"}, true},
		{"missing one", []string{"users:read"}, []string{"users:read", "users:write"}, false},
		{"has none", []string{"orders:read"}, []string{"users:read", "users:write"}, false},
		{"empty required", []string{"users:read"}, []string{}, true},
		{"empty permissions", []string{}, []string{"users:read"}, false},
		{"wildcard has all", []string{"*"}, []string{"users:read", "orders:write"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasAllPermissions(tt.permissions, tt.required); got != tt.want {
				t.Errorf("HasAllPermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}
