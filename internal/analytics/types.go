package analytics

import (
	"time"

	"github.com/google/uuid"
)

type RecentClick struct {
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Referrer  string    `json:"referrer"`
}

type DailyClick struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type Response struct {
	URLID        uuid.UUID        `json:"url_id"`
	TotalClicks  int64            `json:"total_clicks"`
	RecentClicks []RecentClick    `json:"recent_clicks"`
	BrowserStats map[string]int64 `json:"browser_stats"`
	DailyClicks  []DailyClick     `json:"daily_clicks"`
}
