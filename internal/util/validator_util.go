package util

import (
	"errors"
	"regexp"
)

var (
	hasLowerRegex   = regexp.MustCompile(`[a-z]`)
	hasUpperRegex   = regexp.MustCompile(`[A-Z]`)
	hasDigitRegex   = regexp.MustCompile(`[0-9]`)
	hasSpecialRegex = regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`)

	// Check that the domain extension has max 6 characters (e.g., .com, .co.id, .museum)
	tldRegex = regexp.MustCompile(`\.[a-zA-Z]{2,6}$`)
)

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if !hasLowerRegex.MatchString(password) {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasUpperRegex.MatchString(password) {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasDigitRegex.MatchString(password) {
		return errors.New("password must contain at least one number")
	}
	if !hasSpecialRegex.MatchString(password) {
		return errors.New("password must contain at least one special character")
	}
	return nil
}

func ValidateEmailTLD(email string) error {
	if !tldRegex.MatchString(email) {
		return errors.New("invalid email domain extension (max 6 characters)")
	}
	return nil
}
