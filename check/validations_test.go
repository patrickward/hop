package check

import "testing"

func TestValidations(t *testing.T) {
	v := NewValidator()

	username := "jo"
	email := "invalid-email"
	age := 15

	// Simple validations
	v.CheckField(Required(username), "username", "username is required")
	v.CheckField(MinLength(3)(username), "username", "username too short")
	v.CheckField(Email(email), "email", "invalid email format")
	v.CheckField(Min(18)(age), "age", "must be 18 or older")

	// Combining validations
	password := "weak"
	v.CheckField(Required(password), "password", "password is required")
	v.CheckField(MinLength(8)(password), "password", "password too short")
	v.CheckField(Match(`[A-Z]`)(password), "password", "must contain uppercase")
	v.CheckField(Match(`[0-9]`)(password), "password", "must contain number")
}
