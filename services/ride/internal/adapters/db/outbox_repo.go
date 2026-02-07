package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/daffahilmyf/ride-hailing/services/ride/internal/ports/outbound"
)

type OutboxRepo struct {
	DB *gorm.DB
}

func NewOutboxRepo(db *gorm.DB) *OutboxRepo {
	return &OutboxRepo{DB: db}
}

type outboxModel struct {
	ID        string    `gorm:"column:id;primaryKey"`
	Topic     string    `gorm:"column:topic"`
	Payload   string    `gorm:"column:payload"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (outboxModel) TableName() string { return "outbox" }

func (r *OutboxRepo) Enqueue(ctx context.Context, msg outbound.OutboxMessage) error {
	m := outboxModel{
		ID:        msg.ID,
		Topic:     msg.Topic,
		Payload:   msg.Payload,
		CreatedAt: time.Now().UTC(),
	}
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return r.DB.WithContext(ctx).Create(&m).Error
}
