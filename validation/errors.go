package validation

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// ValidationError represents a single validation error for a field or key.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors that can be accumulated.
type ValidationErrors []ValidationError

// Error implements the error interface, combining all error messages.
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// HasErrors returns true if there are any validation errors.
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// Add appends a validation error to the collection.
func (e *ValidationErrors) Add(field, message string) {
	*e = append(*e, ValidationError{Field: field, Message: message})
}

// AddError appends a ValidationError to the collection.
func (e *ValidationErrors) AddError(err ValidationError) {
	*e = append(*e, err)
}

// Merge combines another ValidationErrors into this collection.
func (e *ValidationErrors) Merge(other ValidationErrors) {
	*e = append(*e, other...)
}

// ForField returns all errors for a specific field.
func (e ValidationErrors) ForField(field string) []string {
	var messages []string
	for _, err := range e {
		if err.Field == field {
			messages = append(messages, err.Message)
		}
	}
	return messages
}

// Fields returns all unique field names that have errors.
func (e ValidationErrors) Fields() []string {
	seen := make(map[string]bool)
	var fields []string
	for _, err := range e {
		if err.Field != "" && !seen[err.Field] {
			seen[err.Field] = true
			fields = append(fields, err.Field)
		}
	}
	return fields
}

// Validator is an interface for composable validators.
type Validator interface {
	Validate() ValidationErrors
}

// ValidatorFunc is a function type that implements the Validator interface.
type ValidatorFunc func() ValidationErrors

// Validate implements the Validator interface.
func (f ValidatorFunc) Validate() ValidationErrors {
	return f()
}

// Combine combines multiple validators into a single ValidationErrors result.
func Combine(validators ...Validator) ValidationErrors {
	var errors ValidationErrors
	for _, validator := range validators {
		if validator != nil {
			errors.Merge(validator.Validate())
		}
	}
	return errors
}

// IsRequired checks if a string is not empty.
func IsRequired(value string) bool {
	return strings.TrimSpace(value) != ""
}

// IsRequiredUUID checks if a UUID is not the zero UUID.
func IsRequiredUUID(value uuid.UUID) bool {
	return value != uuid.Nil
}

// MinLength checks if a string has at least the minimum length.
func MinLength(value string, min int) bool {
	return len(value) >= min
}

// MaxLength checks if a string does not exceed the maximum length.
func MaxLength(value string, max int) bool {
	return len(value) <= max
}

// MinValueInt checks if an integer is at least the minimum value.
func MinValueInt(value, min int) bool {
	return value >= min
}

// MaxValueInt checks if an integer does not exceed the maximum value.
func MaxValueInt(value, max int) bool {
	return value <= max
}

// InRange checks if an integer is within the specified range (inclusive).
func InRange(value, min, max int) bool {
	return value >= min && value <= max
}

// OneOf checks if a string is one of the allowed values.
func OneOf(value string, allowed []string) bool {
	for _, a := range allowed {
		if value == a {
			return true
		}
	}
	return false
}

// RequiredString validates that a string field is not empty.
func RequiredString(field, value string) ValidationError {
	if !IsRequired(value) {
		return ValidationError{Field: field, Message: "is required"}
	}
	return ValidationError{}
}

// RequiredUUID validates that a UUID field is not nil.
func RequiredUUID(field string, value uuid.UUID) ValidationError {
	if !IsRequiredUUID(value) {
		return ValidationError{Field: field, Message: "is required"}
	}
	return ValidationError{}
}

// StringMinLength validates that a string has at least the minimum length.
func StringMinLength(field, value string, min int) ValidationError {
	if !MinLength(value, min) {
		return ValidationError{Field: field, Message: fmt.Sprintf("must be at least %d characters", min)}
	}
	return ValidationError{}
}

// StringMaxLength validates that a string does not exceed the maximum length.
func StringMaxLength(field, value string, max int) ValidationError {
	if !MaxLength(value, max) {
		return ValidationError{Field: field, Message: fmt.Sprintf("must be at most %d characters", max)}
	}
	return ValidationError{}
}

// IntMinValue validates that an integer is at least the minimum value.
func IntMinValue(field string, value, min int) ValidationError {
	if !MinValueInt(value, min) {
		return ValidationError{Field: field, Message: fmt.Sprintf("must be at least %d", min)}
	}
	return ValidationError{}
}

// IntMaxValue validates that an integer does not exceed the maximum value.
func IntMaxValue(field string, value, max int) ValidationError {
	if !MaxValueInt(value, max) {
		return ValidationError{Field: field, Message: fmt.Sprintf("must be at most %d", max)}
	}
	return ValidationError{}
}

// IntInRange validates that an integer is within the specified range.
func IntInRange(field string, value, min, max int) ValidationError {
	if !InRange(value, min, max) {
		return ValidationError{Field: field, Message: fmt.Sprintf("must be between %d and %d", min, max)}
	}
	return ValidationError{}
}

// StringOneOf validates that a string is one of the allowed values.
func StringOneOf(field, value string, allowed []string) ValidationError {
	if !OneOf(value, allowed) {
		return ValidationError{Field: field, Message: fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", "))}
	}
	return ValidationError{}
}
