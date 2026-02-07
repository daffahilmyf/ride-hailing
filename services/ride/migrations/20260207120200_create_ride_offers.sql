-- +goose Up
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

-- +goose Down
DROP TABLE IF EXISTS ride_offers;
