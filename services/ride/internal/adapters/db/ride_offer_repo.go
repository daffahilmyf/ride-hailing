package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

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

func (r *RideOfferRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	result := r.DB.WithContext(ctx).Model(&rideOfferModel{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return outbound.ErrNotFound
	}
	return nil
}
