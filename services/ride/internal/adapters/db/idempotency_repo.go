package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IdempotencyRepo struct {
	DB *gorm.DB
}

func NewIdempotencyRepo(db *gorm.DB) *IdempotencyRepo {
	return &IdempotencyRepo{DB: db}
}

type idempotencyModel struct {
	ID           string    `gorm:"column:id;primaryKey"`
	Key          string    `gorm:"column:key"`
	ResponseBody string    `gorm:"column:response_body"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (idempotencyModel) TableName() string { return "idempotency_keys" }

func (r *IdempotencyRepo) Get(ctx context.Context, key string) (string, bool, error) {
	var m idempotencyModel
	err := r.DB.WithContext(ctx).First(&m, "key = ?", key).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", false, nil
		}
		return "", false, err
	}
	return m.ResponseBody, true, nil
}

func (r *IdempotencyRepo) Save(ctx context.Context, key string, response string) error {
	m := idempotencyModel{
		ID:           uuid.NewString(),
		Key:          key,
		ResponseBody: response,
		CreatedAt:    time.Now().UTC(),
	}
	return r.DB.WithContext(ctx).Create(&m).Error
}
