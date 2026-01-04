package auth

import "errors"

var (
	ErrUserNotFound              = errors.New("user not found")
	ErrUserAlreadyExists         = errors.New("user already exists")
	ErrUsernameExists            = errors.New("username already exists")
	ErrInvalidCredentials        = errors.New("invalid credentials")
	ErrInactiveAccount           = errors.New("account is not active")
	ErrRoleNotFound              = errors.New("role not found")
	ErrRoleAlreadyExists         = errors.New("role already exists")
	ErrGrantNotFound             = errors.New("grant not found")
	ErrGrantAlreadyExists        = errors.New("grant already exists")
	ErrPermissionDenied          = errors.New("permission denied")
	ErrInvalidEmail              = errors.New("invalid email")
	ErrInvalidPassword           = errors.New("invalid password")
	ErrInvalidUsername           = errors.New("invalid username")
	ErrInvalidRoleName           = errors.New("invalid role name")
	ErrInvalidDisplayName        = errors.New("invalid display name")
	ErrEncryptionFailed          = errors.New("encryption failed")
	ErrDecryptionFailed          = errors.New("decryption failed")
	ErrPasswordHashFailed        = errors.New("password hash failed")
	ErrTokenGenerationFailed     = errors.New("token generation failed")
	ErrTokenVerificationFailed   = errors.New("token verification failed")
)
