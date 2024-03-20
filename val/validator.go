package val

import (
	"fmt"
	"net/mail"
	"regexp"
)

var (
	isValidUsername = regexp.MustCompile(`^[a-z0-9_]+$`).MatchString
	isValidFullname = regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString
)

func ValidateString(value string, min int, max int) error {
	n := len(value)
	if n < min || n > max {
		return fmt.Errorf("not valid string length")
	}
	return nil
}

func ValidateUsername(value string) error {
	if err := ValidateString(value, 3, 10); err != nil {
		return fmt.Errorf("not valid username length")
	}
	if !isValidUsername(value) {
		return fmt.Errorf("not valid username")
	}
	return nil
}

func ValidatePassword(value string) error {
	return ValidateString(value, 3, 200)
}

func ValidateEmail(value string) error {
	if err := ValidateString(value, 3, 100); err != nil {
		return fmt.Errorf("not valid email length")
	}
	_, err := mail.ParseAddress(value)
	if err != nil {
		return fmt.Errorf("not valid email")
	}
	return nil
}

func ValidateFullname(value string) error {
	if err := ValidateString(value, 3, 100); err != nil {
		return fmt.Errorf("not valid name length")
	}
	if !isValidFullname(value) {
		return fmt.Errorf("not valid name")
	}
	return nil
}
