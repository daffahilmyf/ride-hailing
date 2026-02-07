package db

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type IdempotencyCleanup struct {
	DB *gorm.DB
}

func NewIdempotencyCleanup(db *gorm.DB) *IdempotencyCleanup {
	return &IdempotencyCleanup{DB: db}
}

func (c *IdempotencyCleanup) DeleteBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	result := c.DB.WithContext(ctx).
		Where("created_at < ?", cutoff).
		Delete(&idempotencyModel{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}
