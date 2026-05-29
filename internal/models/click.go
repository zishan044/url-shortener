package models

import (
	"time"

	"github.com/google/uuid"
)

type Click struct {
	ID        uuid.UUID `json:"id"`
	URLID     uuid.UUID `json:"url_id"`
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Referrer  string    `json:"referrer,omitempty"`
}