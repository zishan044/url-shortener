package analytics

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zishan044/url-shortener/internal/models"
)

type Repository interface {
	CreateClick(ctx context.Context, click *models.Click) error
	CreateClicksBatch(ctx context.Context, clicks []*models.Click) error
	URLExists(ctx context.Context, urlID uuid.UUID) (bool, error)
	GetTotalClicks(ctx context.Context, urlID uuid.UUID) (int64, bool, error)
	CountClicks(ctx context.Context, urlID uuid.UUID) (int64, error)
	GetRecentClicks(ctx context.Context, urlID uuid.UUID, limit int) ([]RecentClick, error)
	GetDailyClicks(ctx context.Context, urlID uuid.UUID) ([]DailyClick, error)
	GetUserAgentCounts(ctx context.Context, urlID uuid.UUID) (map[string]int64, error)
	RefreshSummary(ctx context.Context) error
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

func (r *postgresRepository) URLExists(ctx context.Context, urlID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE id = $1)`

	var exists bool
	err := r.db.QueryRow(ctx, query, urlID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *postgresRepository) GetTotalClicks(ctx context.Context, urlID uuid.UUID) (int64, bool, error) {
	query := `
		SELECT total_clicks
		FROM analytics_summary
		WHERE url_id = $1
	`

	var total int64
	err := r.db.QueryRow(ctx, query, urlID).Scan(&total)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}

	return total, true, nil
}

func (r *postgresRepository) CountClicks(ctx context.Context, urlID uuid.UUID) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM clicks
		WHERE url_id = $1
	`

	var total int64
	err := r.db.QueryRow(ctx, query, urlID).Scan(&total)
	return total, err
}

func (r *postgresRepository) GetRecentClicks(ctx context.Context, urlID uuid.UUID, limit int) ([]RecentClick, error) {
	query := `
		SELECT timestamp, COALESCE(ip, ''), COALESCE(user_agent, ''), COALESCE(referrer, '')
		FROM clicks
		WHERE url_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, urlID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clicks := make([]RecentClick, 0, limit)
	for rows.Next() {
		var click RecentClick
		if err := rows.Scan(&click.Timestamp, &click.IP, &click.UserAgent, &click.Referrer); err != nil {
			return nil, err
		}
		clicks = append(clicks, click)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return clicks, nil
}

func (r *postgresRepository) GetDailyClicks(ctx context.Context, urlID uuid.UUID) ([]DailyClick, error) {
	query := `
		SELECT DATE(timestamp AT TIME ZONE 'UTC') AS click_date, COUNT(*) AS click_count
		FROM clicks
		WHERE url_id = $1
		GROUP BY click_date
		ORDER BY click_date DESC
	`

	rows, err := r.db.Query(ctx, query, urlID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dailyClicks := make([]DailyClick, 0)
	for rows.Next() {
		var clickDate time.Time
		var click DailyClick
		if err := rows.Scan(&clickDate, &click.Count); err != nil {
			return nil, err
		}
		click.Date = clickDate.Format(time.DateOnly)
		dailyClicks = append(dailyClicks, click)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return dailyClicks, nil
}

func (r *postgresRepository) GetUserAgentCounts(ctx context.Context, urlID uuid.UUID) (map[string]int64, error) {
	query := `
		SELECT COALESCE(user_agent, ''), COUNT(*)
		FROM clicks
		WHERE url_id = $1
		GROUP BY user_agent
	`

	rows, err := r.db.Query(ctx, query, urlID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int64)
	for rows.Next() {
		var userAgent string
		var count int64
		if err := rows.Scan(&userAgent, &count); err != nil {
			return nil, err
		}
		counts[userAgent] = count
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return counts, nil
}

func (r *postgresRepository) RefreshSummary(ctx context.Context) error {
	query := `
		INSERT INTO analytics_summary (url_id, total_clicks, last_updated)
		SELECT url_id, COUNT(*) AS total_clicks, NOW() AS last_updated
		FROM clicks
		GROUP BY url_id
		ON CONFLICT (url_id)
		DO UPDATE SET
			total_clicks = EXCLUDED.total_clicks,
			last_updated = EXCLUDED.last_updated
	`

	_, err := r.db.Exec(ctx, query)
	return err
}
