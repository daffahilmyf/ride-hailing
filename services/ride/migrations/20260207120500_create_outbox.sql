-- +goose Up
CREATE TABLE IF NOT EXISTS outbox (
  id UUID PRIMARY KEY,
  topic TEXT NOT NULL,
  payload TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'PENDING',
  attempt_count INT NOT NULL DEFAULT 0,
  last_error TEXT,
  available_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS outbox_topic_idx ON outbox (topic);
CREATE INDEX IF NOT EXISTS outbox_created_at_idx ON outbox (created_at);
CREATE INDEX IF NOT EXISTS outbox_status_idx ON outbox (status, available_at, attempt_count);

-- +goose Down
DROP TABLE IF EXISTS outbox;
