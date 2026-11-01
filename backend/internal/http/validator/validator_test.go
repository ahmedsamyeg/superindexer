package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidator(t *testing.T) {
	// Create a test struct with validation tags
	type TestStruct struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"min=0,max=150"`
	}

	t.Run("Valid input should pass validation", func(t *testing.T) {
		testStruct := TestStruct{
			Name:  "John Doe",
			Email: "john@example.com",
			Age:   30,
		}

		// This should not return an error
		err := BindAndValidate(nil, &testStruct)
		assert.NoError(t, err)
	})

	t.Run("Invalid input should fail validation", func(t *testing.T) {
		testStruct := TestStruct{
			Name:  "", // Required field
			Email: "invalid-email", // Invalid email format
			Age:   200, // Too high age
		}

		// This should return validation errors
		err := BindAndValidate(nil, &testStruct)
		assert.Error(t, err)
	})
}