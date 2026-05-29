package analytics

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zishan044/url-shortener/internal/models"
)

type Repository interface {
	CreateClick(ctx context.Context, click *models.Click) error
	CreateClicksBatch(ctx context.Context, clicks []*models.Click) error
}

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreateClick(ctx context.Context, click *models.Click) error {
	query := `
		INSERT INTO clicks (id, url_id, timestamp, ip, user_agent, referrer)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query, click.ID, click.URLID, click.Timestamp, click.IP, click.UserAgent, click.Referrer)
	return err
}

func (r *postgresRepository) CreateClicksBatch(ctx context.Context, clicks []*models.Click) error {
	if len(clicks) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	query := `
		INSERT INTO clicks (id, url_id, timestamp, ip, user_agent, referrer)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	for _, click := range clicks {
		batch.Queue(query, click.ID, click.URLID, click.Timestamp, click.IP, click.UserAgent, click.Referrer)
	}

	results := r.db.SendBatch(ctx, batch)
	defer results.Close()

	for range clicks {
		_, err := results.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}
