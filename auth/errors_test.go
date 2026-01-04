package auth

import (
	"errors"
	"testing"
)

func TestErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"user not found", ErrUserNotFound, "user not found"},
		{"user already exists", ErrUserAlreadyExists, "user already exists"},
		{"username exists", ErrUsernameExists, "username already exists"},
		{"invalid credentials", ErrInvalidCredentials, "invalid credentials"},
		{"inactive account", ErrInactiveAccount, "account is not active"},
		{"role not found", ErrRoleNotFound, "role not found"},
		{"role already exists", ErrRoleAlreadyExists, "role already exists"},
		{"grant not found", ErrGrantNotFound, "grant not found"},
		{"grant already exists", ErrGrantAlreadyExists, "grant already exists"},
		{"permission denied", ErrPermissionDenied, "permission denied"},
		{"invalid email", ErrInvalidEmail, "invalid email"},
		{"invalid password", ErrInvalidPassword, "invalid password"},
		{"invalid username", ErrInvalidUsername, "invalid username"},
		{"invalid role name", ErrInvalidRoleName, "invalid role name"},
		{"invalid display name", ErrInvalidDisplayName, "invalid display name"},
		{"encryption failed", ErrEncryptionFailed, "encryption failed"},
		{"decryption failed", ErrDecryptionFailed, "decryption failed"},
		{"password hash failed", ErrPasswordHashFailed, "password hash failed"},
		{"token generation failed", ErrTokenGenerationFailed, "token generation failed"},
		{"token verification failed", ErrTokenVerificationFailed, "token verification failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("error message = %q, want %q", tt.err.Error(), tt.want)
			}
		})
	}
}

func TestErrorsAreUnique(t *testing.T) {
	allErrors := []error{
		ErrUserNotFound,
		ErrUserAlreadyExists,
		ErrUsernameExists,
		ErrInvalidCredentials,
		ErrInactiveAccount,
		ErrRoleNotFound,
		ErrRoleAlreadyExists,
		ErrGrantNotFound,
		ErrGrantAlreadyExists,
		ErrPermissionDenied,
		ErrInvalidEmail,
		ErrInvalidPassword,
		ErrInvalidUsername,
		ErrInvalidRoleName,
		ErrInvalidDisplayName,
		ErrEncryptionFailed,
		ErrDecryptionFailed,
		ErrPasswordHashFailed,
		ErrTokenGenerationFailed,
		ErrTokenVerificationFailed,
	}

	for i, err1 := range allErrors {
		for j, err2 := range allErrors {
			if i != j && errors.Is(err1, err2) {
				t.Errorf("errors are not unique: %v and %v", err1, err2)
			}
		}
	}
}
