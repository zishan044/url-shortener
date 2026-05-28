package models

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Url struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	OriginalURL string    `json:"original_url"`
	ShortCode   string    `json:"short_code"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (u *Url) Validate() error {
	return u.ValidateOriginalURL()
}

func (u *Url) ValidateOriginalURL() error {
	return validateOriginalURL(u.OriginalURL)
}

func validateOriginalURL(originalURL string) error {
	if strings.TrimSpace(originalURL) == "" {
		return fmt.Errorf("original_url cannot be empty")
	}

	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		return fmt.Errorf("invalid url format: %w", err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("url must include a scheme (http or https)")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("only http and https schemes are supported")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("url must include a valid host")
	}

	return nil
}