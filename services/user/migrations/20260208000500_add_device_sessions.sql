-- +goose Up
ALTER TABLE refresh_tokens
  ADD COLUMN IF NOT EXISTS device_id TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS user_agent TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS ip TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS refresh_tokens_device_idx ON refresh_tokens(device_id);
CREATE UNIQUE INDEX IF NOT EXISTS refresh_tokens_user_device_active_idx
  ON refresh_tokens(user_id, device_id)
  WHERE revoked_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS refresh_tokens_user_device_active_idx;
DROP INDEX IF EXISTS refresh_tokens_device_idx;

ALTER TABLE refresh_tokens
  DROP COLUMN IF EXISTS ip,
  DROP COLUMN IF EXISTS user_agent,
  DROP COLUMN IF EXISTS device_id;
