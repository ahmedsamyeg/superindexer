package validator

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterTagNameFunc(fieldName)
}

// fieldName returns the field name based on the `json` tag by default,
// falling back to the `validate` tag, then the struct field name.
func fieldName(field validator.FieldLevel) string {
	tag := field.Field().Type().String()
	// Try json tag first
	if jsonTag := field.StructField().Tag.Get("json"); jsonTag != "" {
		name := jsonTag
		if idx := index(name, ","); idx != -1 {
			name = name[:idx]
		}
		if name != "" {
			return name
		}
	}
	// Fall back to validate tag
	if valTag := field.StructField().Tag.Get("validate"); valTag != "" {
		return valTag
	}
	return field.Field().Type().String()
}

func index(s, sep string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			return i
		}
	}
	return -1
}

// BindAndValidate binds the request body and validates it.
// Returns an HTTP error if validation fails.
func BindAndValidate(c echo.Context, obj interface{}) error {
	if err := c.Bind(obj); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body").SetInternal(err)
	}

	if err := validate.Struct(obj); err != nil {
		validationErrors := formatValidationErrors(err)
		return echo.NewHTTPError(http.StatusBadRequest, "validation failed").SetInternal(validationErrors)
	}

	return nil
}

// formatValidationErrors converts validator.ValidationErrors to a user-friendly map.
func formatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	if verr, ok := err.(validator.ValidationErrors); ok {
		for _, e := range verr {
			field := e.Field()
			tag := e.Tag()
			msg := validationTagToMessage(field, tag)
			errors[field] = msg
		}
	}
	return errors
}

// validationTagToMessage maps validator tags to human-readable messages.
func validationTagToMessage(field, tag string) string {
	messages := map[string]string{
		"required":  "is required",
		"email":     "must be a valid email address",
		"min":       "is too short (minimum " + tag + " characters)",
		"max":       "is too long (maximum " + tag + " characters)",
		"len":       "must be exactly " + tag + " characters",
		"alpha":     "must contain only letters",
		"alphanum":  "must contain only alphanumeric characters",
		"numeric":   "must be a valid number",
		"uuid":      "must be a valid UUID",
		"url":       "must be a valid URL",
		"dive":      "contains invalid values",
	}
	if msg, ok := messages[tag]; ok {
		return msg
	}
	return "failed validation for tag '" + tag + "'"
}
