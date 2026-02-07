package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/daffahilmyf/ride-hailing/services/ride/internal/ports/outbound"
)

type OutboxRepo struct {
	DB *gorm.DB
}

func NewOutboxRepo(db *gorm.DB) *OutboxRepo {
	return &OutboxRepo{DB: db}
}

type outboxModel struct {
	ID           string    `gorm:"column:id;primaryKey"`
	Topic        string    `gorm:"column:topic"`
	Payload      string    `gorm:"column:payload"`
	Status       string    `gorm:"column:status"`
	AttemptCount int       `gorm:"column:attempt_count"`
	LastError    *string   `gorm:"column:last_error"`
	AvailableAt  time.Time `gorm:"column:available_at"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (outboxModel) TableName() string { return "outbox" }

func (r *OutboxRepo) Enqueue(ctx context.Context, msg outbound.OutboxMessage) error {
	m := outboxModel{
		ID:           msg.ID,
		Topic:        msg.Topic,
		Payload:      msg.Payload,
		Status:       "PENDING",
		AttemptCount: 0,
		AvailableAt:  time.Now().UTC(),
		CreatedAt:    time.Now().UTC(),
	}
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return r.DB.WithContext(ctx).Create(&m).Error
}

func (r *OutboxRepo) Claim(ctx context.Context, limit int, maxAttempts int) ([]outbound.OutboxMessage, error) {
	if limit <= 0 {
		limit = 10
	}
	if maxAttempts <= 0 {
		maxAttempts = 10
	}

	var rows []outboxModel
	tx := r.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	if err := tx.
		Where("status = ? AND attempt_count < ? AND available_at <= ?", "PENDING", maxAttempts, time.Now().UTC()).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Order("created_at").
		Limit(limit).
		Find(&rows).Error; err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	if len(ids) == 0 {
		_ = tx.Commit()
		return nil, nil
	}
	if err := tx.Model(&outboxModel{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"status":        "PROCESSING",
			"attempt_count": gorm.Expr("attempt_count + 1"),
		}).Error; err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	out := make([]outbound.OutboxMessage, 0, len(rows))
	for _, row := range rows {
		out = append(out, outbound.OutboxMessage{
			ID:      row.ID,
			Topic:   row.Topic,
			Payload: row.Payload,
			Attempt: row.AttemptCount + 1,
		})
	}
	return out, nil
}

func (r *OutboxRepo) MarkSent(ctx context.Context, id string) error {
	return r.DB.WithContext(ctx).Model(&outboxModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       "SENT",
			"last_error":   nil,
			"available_at": time.Now().UTC(),
		}).Error
}

func (r *OutboxRepo) MarkFailed(ctx context.Context, id string, reason string, nextAttemptAt time.Time) error {
	if nextAttemptAt.IsZero() {
		nextAttemptAt = time.Now().UTC().Add(5 * time.Second)
	}
	return r.DB.WithContext(ctx).Model(&outboxModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       "PENDING",
			"last_error":   reason,
			"available_at": nextAttemptAt,
		}).Error
}
