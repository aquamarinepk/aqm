package validation_test

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/aquamarinepk/aqm/validation"
)

// Example of basic validation helpers
func ExampleIsRequired() {
	fmt.Println(validation.IsRequired("hello"))
	fmt.Println(validation.IsRequired(""))
	fmt.Println(validation.IsRequired("   "))
	// Output:
	// true
	// false
	// false
}

// Example of accumulating validation errors
func ExampleValidationErrors_Add() {
	var errors validation.ValidationErrors

	// Validate multiple fields
	if !validation.IsRequired("") {
		errors.Add("email", "is required")
	}
	if !validation.MinLength("ab", 3) {
		errors.Add("username", "must be at least 3 characters")
	}

	// Check if there are errors
	if errors.HasErrors() {
		fmt.Println("Validation failed:")
		for _, field := range errors.Fields() {
			for _, msg := range errors.ForField(field) {
				fmt.Printf("  %s: %s\n", field, msg)
			}
		}
	}
	// Output:
	// Validation failed:
	//   email: is required
	//   username: must be at least 3 characters
}

// Example of composable validators
func ExampleCombine() {
	// Define a user registration form
	type RegistrationForm struct {
		Email    string
		Password string
		Age      int
	}

	form := RegistrationForm{
		Email:    "",
		Password: "short",
		Age:      15,
	}

	// Create individual validators
	emailValidator := validation.ValidatorFunc(func() validation.ValidationErrors {
		var errors validation.ValidationErrors
		if err := validation.RequiredString("email", form.Email); err.Field != "" {
			errors.AddError(err)
		}
		if form.Email != "" {
			if err := validation.ValidateEmail(form.Email); err != nil {
				errors.Add("email", err.Error())
			}
		}
		return errors
	})

	passwordValidator := validation.ValidatorFunc(func() validation.ValidationErrors {
		var errors validation.ValidationErrors
		if err := validation.StringMinLength("password", form.Password, 8); err.Field != "" {
			errors.AddError(err)
		}
		return errors
	})

	ageValidator := validation.ValidatorFunc(func() validation.ValidationErrors {
		var errors validation.ValidationErrors
		if err := validation.IntMinValue("age", form.Age, 18); err.Field != "" {
			errors.AddError(err)
		}
		return errors
	})

	// Combine all validators
	errors := validation.Combine(emailValidator, passwordValidator, ageValidator)

	if errors.HasErrors() {
		fmt.Println("Form validation failed:")
		for _, field := range errors.Fields() {
			for _, msg := range errors.ForField(field) {
				fmt.Printf("  %s: %s\n", field, msg)
			}
		}
	}
	// Output:
	// Form validation failed:
	//   email: is required
	//   password: must be at least 8 characters
	//   age: must be at least 18
}

// Example of field-specific error retrieval for frontend
func ExampleValidationErrors_ForField() {
	var errors validation.ValidationErrors
	errors.Add("email", "is required")
	errors.Add("email", "must be a valid email address")
	errors.Add("password", "is too short")

	// Get errors for a specific field (useful for frontend display)
	emailErrors := errors.ForField("email")
	fmt.Println("Email errors:", emailErrors)

	passwordErrors := errors.ForField("password")
	fmt.Println("Password errors:", passwordErrors)

	// Output:
	// Email errors: [is required must be a valid email address]
	// Password errors: [is too short]
}

// Example of reusable validator composition
func ExampleValidator_reusable() {
	// Create a reusable validator for user data
	type User struct {
		ID    uuid.UUID
		Name  string
		Email string
		Role  string
	}

	createUserValidator := func(user User) validation.Validator {
		return validation.ValidatorFunc(func() validation.ValidationErrors {
			var errors validation.ValidationErrors

			// ID validation
			if err := validation.RequiredUUID("id", user.ID); err.Field != "" {
				errors.AddError(err)
			}

			// Name validation
			if err := validation.RequiredString("name", user.Name); err.Field != "" {
				errors.AddError(err)
			}
			if err := validation.StringMinLength("name", user.Name, 3); err.Field != "" {
				errors.AddError(err)
			}
			if err := validation.StringMaxLength("name", user.Name, 100); err.Field != "" {
				errors.AddError(err)
			}

			// Email validation
			if err := validation.RequiredString("email", user.Email); err.Field != "" {
				errors.AddError(err)
			}
			if user.Email != "" {
				if err := validation.ValidateEmail(user.Email); err != nil {
					errors.Add("email", err.Error())
				}
			}

			// Role validation
			if err := validation.StringOneOf("role", user.Role, []string{"admin", "user", "guest"}); err.Field != "" {
				errors.AddError(err)
			}

			return errors
		})
	}

	// Use the validator
	user := User{
		ID:    uuid.Nil,
		Name:  "ab",
		Email: "invalid",
		Role:  "superadmin",
	}

	errors := createUserValidator(user).Validate()

	if errors.HasErrors() {
		fmt.Println("User validation failed:")
		for _, field := range errors.Fields() {
			for _, msg := range errors.ForField(field) {
				fmt.Printf("  %s: %s\n", field, msg)
			}
		}
	}
	// Output:
	// User validation failed:
	//   id: is required
	//   name: must be at least 3 characters
	//   email: invalid email format
	//   role: must be one of: admin, user, guest
}

// Example of conditional validation
func ExampleValidator_conditional() {
	type UpdateProfileForm struct {
		ChangePassword bool
		OldPassword    string
		NewPassword    string
	}

	form := UpdateProfileForm{
		ChangePassword: true,
		OldPassword:    "",
		NewPassword:    "short",
	}

	validator := validation.ValidatorFunc(func() validation.ValidationErrors {
		var errors validation.ValidationErrors

		// Conditional validation - only validate passwords if changing
		if form.ChangePassword {
			if err := validation.RequiredString("old_password", form.OldPassword); err.Field != "" {
				errors.AddError(err)
			}
			if err := validation.RequiredString("new_password", form.NewPassword); err.Field != "" {
				errors.AddError(err)
			}
			if err := validation.StringMinLength("new_password", form.NewPassword, 8); err.Field != "" {
				errors.AddError(err)
			}
		}

		return errors
	})

	errors := validator.Validate()

	if errors.HasErrors() {
		fmt.Println("Profile update validation failed:")
		for _, field := range errors.Fields() {
			for _, msg := range errors.ForField(field) {
				fmt.Printf("  %s: %s\n", field, msg)
			}
		}
	}
	// Output:
	// Profile update validation failed:
	//   old_password: is required
	//   new_password: must be at least 8 characters
}

// Example of merging validation errors from multiple sources
func ExampleValidationErrors_Merge() {
	// Validate basic fields
	var basicErrors validation.ValidationErrors
	basicErrors.Add("email", "is required")
	basicErrors.Add("name", "is required")

	// Validate complex fields separately
	var advancedErrors validation.ValidationErrors
	advancedErrors.Add("password", "must contain uppercase letter")
	advancedErrors.Add("password", "must contain digit")

	// Merge all errors
	var allErrors validation.ValidationErrors
	allErrors.Merge(basicErrors)
	allErrors.Merge(advancedErrors)

	fmt.Printf("Total errors: %d\n", len(allErrors))
	fmt.Printf("Fields with errors: %v\n", allErrors.Fields())
	// Output:
	// Total errors: 4
	// Fields with errors: [email name password]
}

// Example of custom validator helper
func ExampleValidator_custom() {
	// Create a custom validator for a specific business rule
	validateUniqueEmail := func(email string, existingEmails []string) validation.ValidationError {
		for _, existing := range existingEmails {
			if existing == email {
				return validation.ValidationError{
					Field:   "email",
					Message: "email already exists",
				}
			}
		}
		return validation.ValidationError{}
	}

	existingEmails := []string{"user1@example.com", "user2@example.com"}
	newEmail := "user1@example.com"

	var errors validation.ValidationErrors
	if err := validateUniqueEmail(newEmail, existingEmails); err.Field != "" {
		errors.AddError(err)
	}

	if errors.HasErrors() {
		fmt.Println(errors.Error())
	}
	// Output:
	// email: email already exists
}

// Example of validation bypass for testing (fakeable pattern)
func ExampleValidator_bypass() {
	type ValidationMode int

	const (
		StrictMode ValidationMode = iota
		TestMode
	)

	createValidator := func(mode ValidationMode, email string) validation.Validator {
		return validation.ValidatorFunc(func() validation.ValidationErrors {
			var errors validation.ValidationErrors

			// In test mode, skip validation
			if mode == TestMode {
				return errors
			}

			// In strict mode, perform validation
			if err := validation.RequiredString("email", email); err.Field != "" {
				errors.AddError(err)
			}

			return errors
		})
	}

	// Test with strict mode
	strictErrors := createValidator(StrictMode, "").Validate()
	fmt.Printf("Strict mode errors: %d\n", len(strictErrors))

	// Test with test mode (bypassed)
	testErrors := createValidator(TestMode, "").Validate()
	fmt.Printf("Test mode errors: %d\n", len(testErrors))

	// Output:
	// Strict mode errors: 1
	// Test mode errors: 0
}
