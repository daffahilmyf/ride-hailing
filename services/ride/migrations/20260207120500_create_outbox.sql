-- +goose Up
CREATE TABLE IF NOT EXISTS outbox (
  id UUID PRIMARY KEY,
  topic TEXT NOT NULL,
  payload TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS outbox_topic_idx ON outbox (topic);
CREATE INDEX IF NOT EXISTS outbox_created_at_idx ON outbox (created_at);

-- +goose Down
DROP TABLE IF EXISTS outbox;
