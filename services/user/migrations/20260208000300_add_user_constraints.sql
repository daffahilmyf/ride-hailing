-- +goose Up
ALTER TABLE users
  ADD CONSTRAINT users_email_or_phone_chk CHECK (email IS NOT NULL OR phone IS NOT NULL);

ALTER TABLE users
  ADD CONSTRAINT users_role_chk CHECK (role IN ('rider', 'driver'));

CREATE INDEX IF NOT EXISTS users_role_idx ON users(role);

-- +goose Down
DROP INDEX IF EXISTS users_role_idx;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_chk;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_or_phone_chk;
