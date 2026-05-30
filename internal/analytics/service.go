package analytics

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

const recentClicksLimit = 50

var ErrURLNotFound = errors.New("url not found")

type Service interface {
	GetURLAnalytics(ctx context.Context, urlID uuid.UUID) (*Response, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetURLAnalytics(ctx context.Context, urlID uuid.UUID) (*Response, error) {
	exists, err := s.repo.URLExists(ctx, urlID)
	if err != nil {
		return nil, fmt.Errorf("check url exists: %w", err)
	}
	if !exists {
		return nil, ErrURLNotFound
	}

	totalClicks, found, err := s.repo.GetTotalClicks(ctx, urlID)
	if err != nil {
		return nil, fmt.Errorf("get analytics summary: %w", err)
	}
	if !found {
		totalClicks, err = s.repo.CountClicks(ctx, urlID)
		if err != nil {
			return nil, fmt.Errorf("count clicks: %w", err)
		}
	}

	recentClicks, err := s.repo.GetRecentClicks(ctx, urlID, recentClicksLimit)
	if err != nil {
		return nil, fmt.Errorf("get recent clicks: %w", err)
	}

	dailyClicks, err := s.repo.GetDailyClicks(ctx, urlID)
	if err != nil {
		return nil, fmt.Errorf("get daily clicks: %w", err)
	}

	userAgentCounts, err := s.repo.GetUserAgentCounts(ctx, urlID)
	if err != nil {
		return nil, fmt.Errorf("get user agent counts: %w", err)
	}

	browserStats := emptyBrowserStats()
	for userAgent, count := range userAgentCounts {
		browserStats[ClassifyBrowser(userAgent)] += count
	}

	return &Response{
		URLID:        urlID,
		TotalClicks:  totalClicks,
		RecentClicks: recentClicks,
		BrowserStats: browserStats,
		DailyClicks:  dailyClicks,
	}, nil
}
