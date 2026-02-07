-- +goose Up
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

-- +goose Down
DROP TABLE IF EXISTS rides;
