package url

import (
	"context"
	"fmt"

	"github.com/zishan044/url-shortener/internal/models"
)

type Service interface {
	CreateUrl(ctx context.Context, url *models.Url) error
	GetUrlByShortCode(ctx context.Context, shortCode string) (*models.Url, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateUrl(ctx context.Context, url *models.Url) error {
	// Generate a unique short code with retry logic (max 5 attempts)
	var shortCode string
	var err error
	maxRetries := 5

	for attempt := 0; attempt < maxRetries; attempt++ {
		shortCode, err = GenerateShortCode(6, 8)
		if err != nil {
			return fmt.Errorf("failed to generate short code: %w", err)
		}

		exists, err := s.repo.shortCodeExists(ctx, shortCode)
		if err != nil {
			return fmt.Errorf("failed to check short code uniqueness: %w", err)
		}

		if !exists {
			url.ShortCode = shortCode
			break
		}

		if attempt == maxRetries-1 {
			return fmt.Errorf("failed to generate unique short code after %d attempts", maxRetries)
		}
	}

	if err := url.ValidateOriginalURL(); err != nil {
		return err
	}

	return s.repo.createUrl(ctx, url)
}

func (s *service) GetUrlByShortCode(ctx context.Context, shortCode string) (*models.Url, error) {
	return s.repo.getUrlByShortCode(ctx, shortCode)
}

