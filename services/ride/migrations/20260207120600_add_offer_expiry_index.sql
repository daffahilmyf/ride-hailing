-- +goose Up
CREATE INDEX IF NOT EXISTS ride_offers_status_expires_idx ON ride_offers (status, expires_at);

-- +goose Down
DROP INDEX IF EXISTS ride_offers_status_expires_idx;
