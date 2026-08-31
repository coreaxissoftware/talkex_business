package users

import (
	"errors"
	"unicode"
)

// validatePasswordStrength enforces a banking-grade minimum:
//   - length ≥ 12
//   - at least one uppercase, one lowercase, one digit, one symbol
//   - not composed of a single repeated character
//
// The bind tag already enforces min=8; this raises the bar for new
// signups without breaking older accounts that already exist.
func validatePasswordStrength(pw string) error {
	if len(pw) < 12 {
		return errors.New("Password must be at least 12 characters long")
	}
	if len(pw) > 128 {
		return errors.New("Password must be at most 128 characters long")
	}
	var (
		hasUpper, hasLower, hasDigit, hasSymbol bool
		first                                   rune
		allSame                                 = true
	)
	for i, r := range pw {
		if i == 0 {
			first = r
		} else if r != first {
			allSame = false
		}
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}
	if allSame {
		return errors.New("Password must not be a single repeated character")
	}
	if !hasUpper || !hasLower || !hasDigit || !hasSymbol {
		return errors.New("Password must contain uppercase, lowercase, digit, and symbol")
	}
	return nil
}
