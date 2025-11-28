package helpers

import (
	"regexp"
	"unicode"
)

// ValidateEmail checks if email format is valid
func ValidateEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(pattern, email)
	return match
}

// ValidatePassword checks password strength
func ValidatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false
	
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	
	return hasUpper && hasLower && hasDigit && hasSpecial
}

// ValidateUsername checks username format
func ValidateUsername(username string) bool {
	if len(username) < 3 || len(username) > 50 {
		return false
	}
	pattern := `^[a-zA-Z0-9_-]+$`
	match, _ := regexp.MatchString(pattern, username)
	return match
}
