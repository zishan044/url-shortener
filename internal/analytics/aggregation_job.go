package analytics

import (
	"context"
	"log"
	"time"
)

type AggregationJob struct {
	repo     Repository
	interval time.Duration
}

func NewAggregationJob(repo Repository, interval time.Duration) *AggregationJob {
	return &AggregationJob{
		repo:     repo,
		interval: interval,
	}
}

func (j *AggregationJob) Start(ctx context.Context) {
	if err := j.repo.RefreshSummary(ctx); err != nil {
		log.Printf("analytics summary refresh failed: %v", err)
	}

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := j.repo.RefreshSummary(ctx); err != nil {
				log.Printf("analytics summary refresh failed: %v", err)
			}
		}
	}
}
