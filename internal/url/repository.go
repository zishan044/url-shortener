package url

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zishan044/url-shortener/internal/models"
)

type Repository interface {
	createUrl(ctx context.Context, url *models.Url) error
	getUrlByShortCode(ctx context.Context, shortCode string) (*models.Url, error)
	shortCodeExists(ctx context.Context, shortCode string) (bool, error)
}

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return  &postgresRepository{db: db}
}

func (r *postgresRepository) createUrl(ctx context.Context, url *models.Url) error {
	query := `
		INSERT INTO urls (user_id, original_url, short_code, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query, url.UserID, url.OriginalURL, url.ShortCode, url.ExpiresAt).
				Scan(&url.ID, &url.CreatedAt, &url.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *postgresRepository) getUrlByShortCode(ctx context.Context, shortCode string) (*models.Url, error) {
	query := `
		SELECT id, user_id, original_url, short_code, expires_at, created_at, updated_at 
		FROM urls 
		WHERE short_code = $1
	`
	var url models.Url
	err := r.db.QueryRow(ctx, query, shortCode).
				Scan(&url.ID, &url.UserID, &url.OriginalURL, &url.ShortCode, &url.ExpiresAt, &url.CreatedAt, &url.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("url not found")
		}
		return nil, err
	}
	return &url, nil
}

func (r *postgresRepository) shortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, shortCode).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

