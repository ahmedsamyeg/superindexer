package validator

import (
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	// Use the default tag name function which checks json tags
}

// BindAndValidate binds the request body and validates it.
// Returns an HTTP error if validation fails.
func BindAndValidate(c echo.Context, obj interface{}) error {
	if err := c.Bind(obj); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := validate.Struct(obj); err != nil {
		// Convert validation errors to a readable format
		var errors map[string]string = make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			// Get the json tag name or field name
			tagName := err.Tag()
			if jsonTag := err.StructField(); jsonTag != "" {
				// Parse the json tag to get just the field name
				parts := strings.Split(jsonTag, ",")
				tagName = parts[0]
			}
			
			errors[tagName] = err.Error()
		}
		
		return echo.NewHTTPError(http.StatusBadRequest, "validation failed").SetInternal(err)
	}

	return nil
}