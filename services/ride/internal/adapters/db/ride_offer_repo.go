package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/daffahilmyf/ride-hailing/services/ride/internal/domain"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/ports/outbound"
)

type RideOfferRepo struct {
	DB *gorm.DB
}

func NewRideOfferRepo(db *gorm.DB) *RideOfferRepo {
	return &RideOfferRepo{DB: db}
}

type rideOfferModel struct {
	ID        string    `gorm:"column:id;primaryKey"`
	RideID    string    `gorm:"column:ride_id"`
	DriverID  string    `gorm:"column:driver_id"`
	Status    string    `gorm:"column:status"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (rideOfferModel) TableName() string { return "ride_offers" }

func (r *RideOfferRepo) Create(ctx context.Context, offer outbound.RideOffer) error {
	m := rideOfferModel{
		ID:        offer.ID,
		RideID:    offer.RideID,
		DriverID:  offer.DriverID,
		Status:    offer.Status,
		ExpiresAt: time.Unix(offer.ExpiresAt, 0).UTC(),
		CreatedAt: time.Unix(offer.CreatedAt, 0).UTC(),
	}
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return r.DB.WithContext(ctx).Create(&m).Error
}

func (r *RideOfferRepo) Get(ctx context.Context, id string) (outbound.RideOffer, error) {
	var m rideOfferModel
	if err := r.DB.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return outbound.RideOffer{}, outbound.ErrNotFound
		}
		return outbound.RideOffer{}, err
	}
	return outbound.RideOffer{
		ID:        m.ID,
		RideID:    m.RideID,
		DriverID:  m.DriverID,
		Status:    m.Status,
		ExpiresAt: m.ExpiresAt.Unix(),
		CreatedAt: m.CreatedAt.Unix(),
	}, nil
}

func (r *RideOfferRepo) UpdateStatusIfCurrent(ctx context.Context, id string, currentStatus string, nextStatus string) error {
	result := r.DB.WithContext(ctx).Model(&rideOfferModel{}).
		Where("id = ? AND status = ?", id, currentStatus).
		Update("status", nextStatus)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}
	var count int64
	if err := r.DB.WithContext(ctx).Model(&rideOfferModel{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return outbound.ErrNotFound
	}
	return outbound.ErrConflict
}

func (r *RideOfferRepo) ListExpired(ctx context.Context, cutoff int64, limit int) ([]outbound.RideOffer, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []rideOfferModel
	if err := r.DB.WithContext(ctx).
		Where("status = ? AND expires_at <= ?", string(domain.OfferPending), time.Unix(cutoff, 0).UTC()).
		Order("expires_at").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]outbound.RideOffer, 0, len(rows))
	for _, row := range rows {
		out = append(out, outbound.RideOffer{
			ID:        row.ID,
			RideID:    row.RideID,
			DriverID:  row.DriverID,
			Status:    row.Status,
			ExpiresAt: row.ExpiresAt.Unix(),
			CreatedAt: row.CreatedAt.Unix(),
		})
	}
	return out, nil
}
