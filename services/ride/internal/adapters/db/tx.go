package db

import (
	"gorm.io/gorm"

	"github.com/daffahilmyf/ride-hailing/services/ride/internal/ports/outbound"
)

type TxManager struct {
	DB *gorm.DB
}

func NewTxManager(db *gorm.DB) *TxManager {
	return &TxManager{DB: db}
}

func (m *TxManager) Begin() (outbound.Tx, error) {
	tx := m.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &gormTx{tx: tx}, nil
}

type gormTx struct {
	tx *gorm.DB
}

func (t *gormTx) Commit() error {
	return t.tx.Commit().Error
}

func (t *gormTx) Rollback() error {
	return t.tx.Rollback().Error
}

func (t *gormTx) RideRepo() outbound.RideRepo {
	return NewRideRepo(t.tx)
}

func (t *gormTx) IdempotencyRepo() outbound.IdempotencyRepo {
	return NewIdempotencyRepo(t.tx)
}
