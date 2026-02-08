-- +goose Up
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS phone_verified_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS failed_login_count INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS locked_until TIMESTAMPTZ;

ALTER TABLE rider_profiles
  ADD COLUMN IF NOT EXISTS rating DOUBLE PRECISION NOT NULL DEFAULT 5.0,
  ADD COLUMN IF NOT EXISTS preferred_language TEXT NOT NULL DEFAULT 'en';

ALTER TABLE driver_profiles
  ADD COLUMN IF NOT EXISTS vehicle_make TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS vehicle_plate TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS license_number TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS verified BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS rating DOUBLE PRECISION NOT NULL DEFAULT 5.0;

CREATE TABLE IF NOT EXISTS verification_codes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  channel TEXT NOT NULL,
  code_hash TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  consumed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS verification_codes_channel_idx ON verification_codes(channel);
CREATE INDEX IF NOT EXISTS verification_codes_user_idx ON verification_codes(user_id);

-- +goose Down
DROP INDEX IF EXISTS verification_codes_user_idx;
DROP INDEX IF EXISTS verification_codes_channel_idx;
DROP TABLE IF EXISTS verification_codes;

ALTER TABLE driver_profiles
  DROP COLUMN IF EXISTS rating,
  DROP COLUMN IF EXISTS verified,
  DROP COLUMN IF EXISTS license_number,
  DROP COLUMN IF EXISTS vehicle_plate,
  DROP COLUMN IF EXISTS vehicle_make;

ALTER TABLE rider_profiles
  DROP COLUMN IF EXISTS preferred_language,
  DROP COLUMN IF EXISTS rating;

ALTER TABLE users
  DROP COLUMN IF EXISTS locked_until,
  DROP COLUMN IF EXISTS failed_login_count,
  DROP COLUMN IF EXISTS phone_verified_at,
  DROP COLUMN IF EXISTS email_verified_at;
