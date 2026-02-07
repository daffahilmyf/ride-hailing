DB_DSN ?= postgres://ride:ride@localhost:5432/rides?sslmode=disable
MIGRATIONS_DIR ?= ./migrations

.PHONY: migrate-up migrate-down

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" down
