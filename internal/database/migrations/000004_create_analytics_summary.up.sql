CREATE TABLE IF NOT EXISTS analytics_summary (
    url_id UUID PRIMARY KEY REFERENCES urls(id) ON DELETE CASCADE,
    total_clicks BIGINT NOT NULL DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_clicks_url_id_timestamp_desc
    ON clicks(url_id, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_clicks_url_id_user_agent
    ON clicks(url_id, user_agent);
