package domain

import (
	"regexp"
	"strings"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

type Email string

func ValidateEmail(raw string) (Email, error) {
	email := strings.TrimSpace(strings.ToLower(raw))
	if email == "" {
		return "", ErrInvalidEmail
	}
	if !emailRegex.MatchString(email) {
		return "", ErrInvalidEmail
	}
	return Email(email), nil
}

func (e Email) String() string {
	return string(e)
}