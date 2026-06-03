package validators

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

func ValidateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("email cannot be empty")
	}

	if len(email) > 254 {
		return fmt.Errorf("email is too long")
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid email format")
	}

	local, domain := parts[0], parts[1]

	if len(local) == 0 || len(domain) == 0 {
		return fmt.Errorf("invalid email format")
	}

	if local[0] == '.' || local[len(local)-1] == '.' {
		return fmt.Errorf("email local part cannot start or end with a dot")
	}

	if strings.Contains(local, "..") {
		return fmt.Errorf("email local part cannot contain consecutive dots")
	}

	if !strings.Contains(domain, ".") {
		return fmt.Errorf("email domain must contain a dot")
	}

	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return fmt.Errorf("password is too long")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return fmt.Errorf("password must contain uppercase, lowercase, and digit characters")
	}

	return nil
}

func ValidateUsername(username string) error {
	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}

	if len(username) > 32 {
		return fmt.Errorf("username must be at most 32 characters long")
	}

	for _, char := range username {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' && char != '-' {
			return fmt.Errorf("username can only contain letters, numbers, underscores, and hyphens")
		}
	}

	return nil
}

func SanitizeURL(originalURL string) (string, error) {
	if strings.TrimSpace(originalURL) == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme == "" {
		return "", fmt.Errorf("URL must include a scheme (http or https)")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", fmt.Errorf("only http and https schemes are supported")
	}

	if parsedURL.Host == "" {
		return "", fmt.Errorf("URL must include a valid host")
	}

	normalizedURL := parsedURL.String()

	if len(normalizedURL) > 2048 {
		return "", fmt.Errorf("URL is too long (max 2048 characters)")
	}

	return normalizedURL, nil
}

func ValidateShortCode(shortCode string) error {
	if len(shortCode) < 3 {
		return fmt.Errorf("short code must be at least 3 characters long")
	}

	if len(shortCode) > 20 {
		return fmt.Errorf("short code must be at most 20 characters long")
	}

	for _, char := range shortCode {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' && char != '-' {
			return fmt.Errorf("short code can only contain letters, numbers, underscores, and hyphens")
		}
	}

	return nil
}
