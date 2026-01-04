package auth

import (
	"strings"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"valid email with subdomain", "user@mail.example.com", false},
		{"empty email", "", true},
		{"missing @", "testexample.com", true},
		{"missing domain", "test@", true},
		{"missing local part", "@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name    string
		pass    string
		wantErr bool
	}{
		{"valid password", "Test1234!", false},
		{"too short", "Test1!", true},
		{"no uppercase", "test1234!", true},
		{"no lowercase", "TEST1234!", true},
		{"no digit", "TestTest!", true},
		{"no special", "Test1234", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.pass)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name    string
		username string
		wantErr bool
	}{
		{"valid username", "john_doe", false},
		{"valid with dash", "john-doe", false},
		{"valid with dot", "john.doe", false},
		{"valid with numbers", "john123", false},
		{"too short", "ab", true},
		{"too long", "a" + strings.Repeat("b", 32), true},
		{"invalid chars space", "john doe", true},
		{"invalid chars special", "john@doe", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDisplayName(t *testing.T) {
	tests := []struct {
		name    string
		dispName string
		wantErr bool
	}{
		{"valid name", "John Doe", false},
		{"valid with unicode", "João Silva", false},
		{"valid single char", "A", false},
		{"empty", "", true},
		{"only spaces", "   ", true},
		{"too long", strings.Repeat("a", 129), true},
		{"max length valid", strings.Repeat("a", 128), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDisplayName(tt.dispName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDisplayName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRoleName(t *testing.T) {
	tests := []struct {
		name     string
		roleName string
		wantErr  bool
	}{
		{"valid role", "admin", false},
		{"valid with underscore", "super_admin", false},
		{"valid with dash", "super-admin", false},
		{"valid with numbers", "admin123", false},
		{"too short", "a", true},
		{"too long", strings.Repeat("a", 65), true},
		{"invalid chars space", "super admin", true},
		{"invalid chars special", "admin@role", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRoleName(tt.roleName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRoleName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  string
	}{
		{"lowercase", "Test@Example.COM", "test@example.com"},
		{"trim spaces", "  test@example.com  ", "test@example.com"},
		{"already normalized", "test@example.com", "test@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeEmail(tt.email); got != tt.want {
				t.Errorf("NormalizeEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     string
	}{
		{"lowercase", "JohnDoe", "johndoe"},
		{"trim spaces", "  johndoe  ", "johndoe"},
		{"trim leading dots", "..johndoe", "johndoe"},
		{"trim trailing dots", "johndoe..", "johndoe"},
		{"trim leading dash", "-johndoe", "johndoe"},
		{"trim trailing underscore", "johndoe_", "johndoe"},
		{"combined", "  .JohnDoe_.  ", "johndoe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeUsername(tt.username); got != tt.want {
				t.Errorf("NormalizeUsername() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		dispName string
		want     string
	}{
		{"trim spaces", "  John Doe  ", "John Doe"},
		{"preserve case", "John Doe", "John Doe"},
		{"already normalized", "John", "John"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeDisplayName(tt.dispName); got != tt.want {
				t.Errorf("NormalizeDisplayName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeRoleName(t *testing.T) {
	tests := []struct {
		name     string
		roleName string
		want     string
	}{
		{"lowercase", "SuperAdmin", "superadmin"},
		{"trim spaces", "  admin  ", "admin"},
		{"combined", "  SuperAdmin  ", "superadmin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeRoleName(tt.roleName); got != tt.want {
				t.Errorf("NormalizeRoleName() = %v, want %v", got, tt.want)
			}
		})
	}
}
