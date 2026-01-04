package service

import (
	"context"
	"fmt"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/google/uuid"
)

const SuperadminEmail = "superadmin@system.local"

// SignUp creates a new user with email and password
func SignUp(ctx context.Context, store auth.UserStore, crypto CryptoService, email, password, username, displayName string) (*auth.User, error) {
	if store == nil {
		return nil, fmt.Errorf("user store is required")
	}
	if crypto == nil {
		return nil, fmt.Errorf("crypto service is required")
	}

	// Validate and normalize inputs
	email = auth.NormalizeEmail(email)
	if err := auth.ValidateEmail(email); err != nil {
		return nil, err
	}

	username = auth.NormalizeUsername(username)
	if err := auth.ValidateUsername(username); err != nil {
		return nil, err
	}

	displayName = auth.NormalizeDisplayName(displayName)
	if err := auth.ValidateDisplayName(displayName); err != nil {
		return nil, err
	}

	if err := auth.ValidatePassword(password); err != nil {
		return nil, err
	}

	// Check if user exists by email
	emailLookup := crypto.ComputeLookupHash(email)
	existing, err := store.GetByEmailLookup(ctx, emailLookup)
	if err != nil && err != auth.ErrUserNotFound {
		return nil, fmt.Errorf("check existing user: %w", err)
	}
	if existing != nil {
		return nil, auth.ErrUserAlreadyExists
	}

	// Check if username exists
	existingUsername, err := store.GetByUsername(ctx, username)
	if err != nil && err != auth.ErrUserNotFound {
		return nil, fmt.Errorf("check existing username: %w", err)
	}
	if existingUsername != nil {
		return nil, auth.ErrUsernameExists
	}

	// Create user
	user := auth.NewUser()
	user.Username = username
	user.Name = displayName
	user.Status = auth.UserStatusActive

	if err := user.SetEmail(email, crypto.EncryptionKey(), crypto.SigningKey()); err != nil {
		return nil, fmt.Errorf("encrypt email: %w", err)
	}

	if err := user.SetPassword(password); err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user.BeforeCreate()

	if err := store.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

// SignIn authenticates a user with email and password
func SignIn(ctx context.Context, store auth.UserStore, crypto CryptoService, tokenGen TokenGenerator, email, password string) (*auth.User, string, error) {
	if store == nil {
		return nil, "", fmt.Errorf("user store is required")
	}
	if crypto == nil {
		return nil, "", fmt.Errorf("crypto service is required")
	}
	if tokenGen == nil {
		return nil, "", fmt.Errorf("token generator is required")
	}

	email = auth.NormalizeEmail(email)
	emailLookup := crypto.ComputeLookupHash(email)

	user, err := store.GetByEmailLookup(ctx, emailLookup)
	if err == auth.ErrUserNotFound || user == nil {
		return nil, "", auth.ErrInvalidCredentials
	}
	if err != nil {
		return nil, "", fmt.Errorf("lookup user: %w", err)
	}

	if !user.VerifyPassword(password) {
		return nil, "", auth.ErrInvalidCredentials
	}

	if user.Status != auth.UserStatusActive {
		return nil, "", auth.ErrInactiveAccount
	}

	token, err := tokenGen.GenerateToken(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("generate token: %w", err)
	}

	return user, token, nil
}

// SignInByPIN authenticates a user using their PIN for lightweight access
func SignInByPIN(ctx context.Context, store auth.UserStore, crypto CryptoService, pin string) (*auth.User, error) {
	if store == nil {
		return nil, fmt.Errorf("user store is required")
	}
	if crypto == nil {
		return nil, fmt.Errorf("crypto service is required")
	}

	if len(pin) < 4 || len(pin) > 8 {
		return nil, auth.ErrInvalidCredentials
	}

	pinLookup := crypto.ComputePINLookupHash(pin)

	user, err := store.GetByPINLookup(ctx, pinLookup)
	if err == auth.ErrUserNotFound || user == nil {
		return nil, auth.ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("lookup user by PIN: %w", err)
	}

	if !user.VerifyPIN(pin, crypto.EncryptionKey()) {
		return nil, auth.ErrInvalidCredentials
	}

	if user.Status != auth.UserStatusActive {
		return nil, auth.ErrInactiveAccount
	}

	return user, nil
}

// Bootstrap creates the initial superadmin user with a one-time password
func Bootstrap(ctx context.Context, store auth.UserStore, crypto CryptoService, pwdGen PasswordGenerator) (*auth.User, string, error) {
	if store == nil {
		return nil, "", fmt.Errorf("user store is required")
	}
	if crypto == nil {
		return nil, "", fmt.Errorf("crypto service is required")
	}
	if pwdGen == nil {
		return nil, "", fmt.Errorf("password generator is required")
	}

	emailLookup := crypto.ComputeLookupHash(SuperadminEmail)

	// Check if superadmin already exists
	existing, err := store.GetByEmailLookup(ctx, emailLookup)
	if err != nil && err != auth.ErrUserNotFound {
		return nil, "", fmt.Errorf("lookup superadmin: %w", err)
	}
	if existing != nil {
		return existing, "", nil
	}

	// Create superadmin
	user := auth.NewUser()
	user.Username = "superadmin"
	user.Name = "Super Administrator"
	user.Status = auth.UserStatusActive

	if err := user.SetEmail(SuperadminEmail, crypto.EncryptionKey(), crypto.SigningKey()); err != nil {
		return nil, "", fmt.Errorf("encrypt email: %w", err)
	}

	password := pwdGen.GeneratePassword()
	if err := user.SetPassword(password); err != nil {
		return nil, "", fmt.Errorf("hash password: %w", err)
	}

	user.BeforeCreate()

	if err := store.Create(ctx, user); err != nil {
		return nil, "", fmt.Errorf("create superadmin: %w", err)
	}

	return user, password, nil
}

// GeneratePIN creates a unique PIN for a user (for lightweight authentication)
func GeneratePIN(ctx context.Context, store auth.UserStore, crypto CryptoService, pinGen PINGenerator, user *auth.User) (string, error) {
	if store == nil {
		return "", fmt.Errorf("user store is required")
	}
	if crypto == nil {
		return "", fmt.Errorf("crypto service is required")
	}
	if pinGen == nil {
		return "", fmt.Errorf("PIN generator is required")
	}
	if user == nil {
		return "", fmt.Errorf("user is required")
	}

	const maxAttempts = 10
	for attempt := 0; attempt < maxAttempts; attempt++ {
		pin := pinGen.GeneratePIN()
		pinLookup := crypto.ComputePINLookupHash(pin)

		// Check for collision
		existing, err := store.GetByPINLookup(ctx, pinLookup)
		if err != nil && err != auth.ErrUserNotFound {
			return "", fmt.Errorf("check PIN collision: %w", err)
		}
		if existing != nil {
			continue
		}

		// Set PIN on user
		if err := user.SetPIN(pin, crypto.EncryptionKey(), crypto.SigningKey()); err != nil {
			return "", fmt.Errorf("encrypt PIN: %w", err)
		}

		user.BeforeUpdate()

		if err := store.Update(ctx, user); err != nil {
			return "", fmt.Errorf("update user with PIN: %w", err)
		}

		return pin, nil
	}

	return "", fmt.Errorf("failed to generate unique PIN after %d attempts", maxAttempts)
}

// GetUserByID retrieves a user by their ID
func GetUserByID(ctx context.Context, store auth.UserStore, id uuid.UUID) (*auth.User, error) {
	if store == nil {
		return nil, fmt.Errorf("user store is required")
	}
	return store.Get(ctx, id)
}

// GetUserByUsername retrieves a user by their username
func GetUserByUsername(ctx context.Context, store auth.UserStore, username string) (*auth.User, error) {
	if store == nil {
		return nil, fmt.Errorf("user store is required")
	}
	username = auth.NormalizeUsername(username)
	return store.GetByUsername(ctx, username)
}

// ListUsers retrieves all users
func ListUsers(ctx context.Context, store auth.UserStore) ([]*auth.User, error) {
	if store == nil {
		return nil, fmt.Errorf("user store is required")
	}
	return store.List(ctx)
}

// ListUsersByStatus retrieves users by status
func ListUsersByStatus(ctx context.Context, store auth.UserStore, status auth.UserStatus) ([]*auth.User, error) {
	if store == nil {
		return nil, fmt.Errorf("user store is required")
	}
	return store.ListByStatus(ctx, status)
}

// UpdateUser updates a user's information
func UpdateUser(ctx context.Context, store auth.UserStore, user *auth.User) error {
	if store == nil {
		return fmt.Errorf("user store is required")
	}
	if user == nil {
		return fmt.Errorf("user is required")
	}
	user.BeforeUpdate()
	return store.Update(ctx, user)
}

// DeleteUser soft-deletes a user
func DeleteUser(ctx context.Context, store auth.UserStore, id uuid.UUID) error {
	if store == nil {
		return fmt.Errorf("user store is required")
	}
	return store.Delete(ctx, id)
}
