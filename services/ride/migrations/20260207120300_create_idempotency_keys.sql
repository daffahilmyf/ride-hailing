-- +goose Up
CREATE TABLE IF NOT EXISTS idempotency_keys (
  id UUID PRIMARY KEY,
  key TEXT NOT NULL,
  response_body JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idempotency_keys_key_idx ON idempotency_keys (key);

-- +goose Down
DROP TABLE IF EXISTS idempotency_keys;
