-- +goose Up
ALTER TABLE ride_offers
  ADD CONSTRAINT ride_offers_ride_fk
  FOREIGN KEY (ride_id) REFERENCES rides(id) ON DELETE CASCADE;

ALTER TABLE ride_offers
  ADD CONSTRAINT ride_offers_driver_id_not_null CHECK (driver_id IS NOT NULL);

-- +goose Down
ALTER TABLE ride_offers DROP CONSTRAINT IF EXISTS ride_offers_driver_id_not_null;
ALTER TABLE ride_offers DROP CONSTRAINT IF EXISTS ride_offers_ride_fk;
