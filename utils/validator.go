package utils

import (
	"regexp"
	"strings"
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,50}$`)
)

func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func ValidateUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

func ValidatePassword(password string) (bool, string) {
	if len(password) < 6 {
		return false, "Password must be at least 6 characters"
	}
	if len(password) > 100 {
		return false, "Password must not exceed 100 characters"
	}
	return true, ""
}

func ValidateRequired(value, fieldName string) (bool, string) {
	if strings.TrimSpace(value) == "" {
		return false, fieldName + " is required"
	}
	return true, ""
}

func ValidateMinLength(value, fieldName string, minLen int) (bool, string) {
	if len(value) < minLen {
		return false, fieldName + " must be at least " + string(rune(minLen+'0')) + " characters"
	}
	return true, ""
}

func ValidateMaxLength(value, fieldName string, maxLen int) (bool, string) {
	if len(value) > maxLen {
		return false, fieldName + " must not exceed " + string(rune(maxLen+'0')) + " characters"
	}
	return true, ""
}
