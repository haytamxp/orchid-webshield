package validation

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"unicode"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{3,32}$`)

// Username vérifie le nom d'utilisateur.
func Username(s string) error {
	if !usernamePattern.MatchString(s) {
		return errors.New("username must be 3 to 32 characters: letters, digits, '_', '.' or '-'")
	}
	return nil
}

// Email vérifie le format de l'adresse et la normalise (minuscules, sans espaces).
func Email(s string) (string, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if len(s) > 254 {
		return "", errors.New("email is too long")
	}
	addr, err := mail.ParseAddress(s)
	if err != nil || addr.Address != s {
		return "", errors.New("email is not valid")
	}
	return s, nil
}

// Password applique la politique de mot de passe.
// La limite haute de 72 octets est celle de bcrypt.
func Password(s string) error {
	if len(s) < 10 {
		return errors.New("password must be at least 10 characters")
	}
	if len(s) > 72 {
		return errors.New("password must be at most 72 bytes")
	}
	var hasLetter, hasDigit bool
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return errors.New("password must contain at least one letter and one digit")
	}
	return nil
}
