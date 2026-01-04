package auth

import (
	"errors"
	"strings"
	"unicode"

	"github.com/aquamarinepk/aqm/validation"
)

func ValidateEmail(email string) error {
	err := validation.ValidateEmail(email)
	if err != nil {
		// Convert validation errors to auth domain errors
		if errors.Is(err, validation.ErrEmailInvalid) ||
			errors.Is(err, validation.ErrEmailTooLong) ||
			errors.Is(err, validation.ErrEmailLocalTooLong) {
			return ErrInvalidEmail
		}
		return err
	}
	return nil
}

func ValidatePassword(password string) error {
	err := validation.ValidatePassword(password)
	if err != nil {
		// Convert validation errors to auth domain errors
		if errors.Is(err, validation.ErrPasswordTooShort) ||
			errors.Is(err, validation.ErrPasswordTooLong) ||
			errors.Is(err, validation.ErrPasswordNoUppercase) ||
			errors.Is(err, validation.ErrPasswordNoLowercase) ||
			errors.Is(err, validation.ErrPasswordNoDigit) ||
			errors.Is(err, validation.ErrPasswordNoSpecial) {
			return ErrInvalidPassword
		}
		return err
	}
	return nil
}

func ValidateUsername(username string) error {
	if len(username) < 3 {
		return ErrInvalidUsername
	}
	if len(username) > 32 {
		return ErrInvalidUsername
	}

	for _, r := range username {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' && r != '.' {
			return ErrInvalidUsername
		}
	}

	return nil
}

func ValidateDisplayName(name string) error {
	trimmed := strings.TrimSpace(name)
	if len(trimmed) < 1 {
		return ErrInvalidDisplayName
	}
	if len(trimmed) > 128 {
		return ErrInvalidDisplayName
	}
	return nil
}

func ValidateRoleName(name string) error {
	if len(name) < 2 {
		return ErrInvalidRoleName
	}
	if len(name) > 64 {
		return ErrInvalidRoleName
	}

	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return ErrInvalidRoleName
		}
	}

	return nil
}

func NormalizeEmail(email string) string {
	return validation.NormalizeEmail(email)
}

func NormalizeUsername(username string) string {
	trimmed := strings.ToLower(strings.TrimSpace(username))
	return strings.Trim(trimmed, "._-")
}

func NormalizeDisplayName(name string) string {
	return strings.TrimSpace(name)
}

func NormalizeRoleName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
