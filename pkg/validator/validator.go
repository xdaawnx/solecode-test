package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps go-playground/validator
type Validator struct {
	validate *validator.Validate
}

// New creates a new validator instance
func New() *Validator {
	validate := validator.New()

	// Register custom validations
	registerCustomValidations(validate)

	// Use JSON tag names for field names
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{
		validate: validate,
	}
}

// registerCustomValidations registers custom validation functions
func registerCustomValidations(validate *validator.Validate) {
	// Register custom name validation
	_ = validate.RegisterValidation("name", func(fl validator.FieldLevel) bool {
		name, ok := fl.Field().Interface().(string)
		if !ok {
			return false
		}
		return isValidName(name)
	})

	// Register custom password validation
	_ = validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		password, ok := fl.Field().Interface().(string)
		if !ok {
			return false
		}
		return isValidPassword(password)
	})
}

// ValidateStruct validates a struct against its validation tags
func (v *Validator) ValidateStruct(s interface{}) error {
	if err := v.validate.Struct(s); err != nil {
		return convertValidationError(err)
	}
	return nil
}

// ValidateVar validates a single variable against a tag
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	if err := v.validate.Var(field, tag); err != nil {
		return convertValidationError(err)
	}
	return nil
}

// ValidationError represents a structured validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"-"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return "no validation errors"
	}

	var messages []string
	for _, err := range v {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

// convertValidationError converts validator.ValidationErrors to our custom format
func convertValidationError(err error) ValidationErrors {
	var validationErrors ValidationErrors

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range ve {
			validationError := ValidationError{
				Field: fieldError.Field(),
				Tag:   fieldError.Tag(),
				Value: fieldError.Param(),
			}

			// Customize error messages based on tag and field
			validationError.Message = getErrorMessage(fieldError.Field(), fieldError.Tag(), fieldError.Param())

			validationErrors = append(validationErrors, validationError)
		}
	}

	return validationErrors
}

// getErrorMessage returns user-friendly error messages
func getErrorMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return "must be a valid email address"
	case "name":
		return "must contain only letters, spaces, hyphens, and apostrophes"
	case "min":
		return fmt.Sprintf("must be at least %s characters long", param)
	case "max":
		return fmt.Sprintf("must be at most %s characters long", param)
	case "len":
		return fmt.Sprintf("must be exactly %s characters long", param)
	case "password":
		return "must contain at least 8 characters including uppercase, lowercase, number and special character"
	case "numeric":
		return "must be a valid number"
	case "alphanum":
		return "must contain only letters and numbers"
	default:
		return fmt.Sprintf("failed %s validation", tag)
	}
}

// Helper validation functions
func isValidName(name string) bool {
	if name == "" {
		return false
	}

	// Name should contain only letters, spaces, hyphens, and apostrophes
	nameRegex := `^[a-zA-Z\s\-'\.]+$`
	matched, _ := regexp.MatchString(nameRegex, name)
	return matched
}

func isValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	// At least one uppercase letter
	upperCase := `[A-Z]`
	// At least one lowercase letter
	lowerCase := `[a-z]`
	// At least one number
	number := `[0-9]`
	// At least one special character
	specialChar := `[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`

	return regexp.MustCompile(upperCase).MatchString(password) &&
		regexp.MustCompile(lowerCase).MatchString(password) &&
		regexp.MustCompile(number).MatchString(password) &&
		regexp.MustCompile(specialChar).MatchString(password)
}

// Convenience methods for common validations
func (v *Validator) ValidateEmail(email string) error {
	return v.ValidateVar(email, "required,email")
}

func (v *Validator) ValidateName(name string) error {
	return v.ValidateVar(name, "required,min=2,max=100,name")
}

func (v *Validator) ValidateID(id int64) error {
	return v.ValidateVar(id, "required,min=1")
}
