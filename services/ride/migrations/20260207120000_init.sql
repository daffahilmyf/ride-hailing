-- +goose Up
-- rides
CREATE TABLE IF NOT EXISTS rides (
  id UUID PRIMARY KEY,
  rider_id UUID NOT NULL,
  driver_id UUID NULL,
  status TEXT NOT NULL,
  pickup_lat DOUBLE PRECISION NOT NULL,
  pickup_lng DOUBLE PRECISION NOT NULL,
  dropoff_lat DOUBLE PRECISION NOT NULL,
  dropoff_lng DOUBLE PRECISION NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS rides_rider_status_idx ON rides (rider_id, status);
CREATE INDEX IF NOT EXISTS rides_status_idx ON rides (status);
CREATE INDEX IF NOT EXISTS rides_driver_status_idx ON rides (driver_id, status);

-- ride_offers
CREATE TABLE IF NOT EXISTS ride_offers (
  id UUID PRIMARY KEY,
  ride_id UUID NOT NULL,
  driver_id UUID NOT NULL,
  status TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS ride_offers_ride_idx ON ride_offers (ride_id);
CREATE INDEX IF NOT EXISTS ride_offers_driver_status_idx ON ride_offers (driver_id, status);

-- idempotency_keys
CREATE TABLE IF NOT EXISTS idempotency_keys (
  id UUID PRIMARY KEY,
  key TEXT NOT NULL,
  response_body JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idempotency_keys_key_idx ON idempotency_keys (key);

-- +goose Down
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS ride_offers;
DROP TABLE IF EXISTS rides;
