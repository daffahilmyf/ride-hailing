package db

import (
	"context"
	"time"

	"github.com/daffahilmyf/ride-hailing/services/ride/internal/ports/outbound"
	"gorm.io/gorm"
)

type RideRepo struct {
	DB *gorm.DB
}

func NewRideRepo(db *gorm.DB) *RideRepo {
	return &RideRepo{DB: db}
}

type rideModel struct {
	ID         string    `gorm:"column:id;primaryKey"`
	RiderID    string    `gorm:"column:rider_id"`
	DriverID   *string   `gorm:"column:driver_id"`
	Status     string    `gorm:"column:status"`
	PickupLat  float64   `gorm:"column:pickup_lat"`
	PickupLng  float64   `gorm:"column:pickup_lng"`
	DropoffLat float64   `gorm:"column:dropoff_lat"`
	DropoffLng float64   `gorm:"column:dropoff_lng"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (rideModel) TableName() string {
	return "rides"
}

func (r *RideRepo) Create(ctx context.Context, ride outbound.Ride) error {
	m := rideModel{
		ID:         ride.ID,
		RiderID:    ride.RiderID,
		DriverID:   ride.DriverID,
		Status:     ride.Status,
		PickupLat:  ride.PickupLat,
		PickupLng:  ride.PickupLng,
		DropoffLat: ride.DropoffLat,
		DropoffLng: ride.DropoffLng,
		CreatedAt:  ride.CreatedAt,
		UpdatedAt:  ride.UpdatedAt,
	}
	return r.DB.WithContext(ctx).Create(&m).Error
}

func (r *RideRepo) Get(ctx context.Context, id string) (outbound.Ride, error) {
	var m rideModel
	if err := r.DB.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return outbound.Ride{}, err
	}
	return outbound.Ride{
		ID:         m.ID,
		RiderID:    m.RiderID,
		DriverID:   m.DriverID,
		Status:     m.Status,
		PickupLat:  m.PickupLat,
		PickupLng:  m.PickupLng,
		DropoffLat: m.DropoffLat,
		DropoffLng: m.DropoffLng,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}, nil
}

func (r *RideRepo) UpdateStatus(ctx context.Context, id string, status string, updatedAt time.Time) error {
	return r.DB.WithContext(ctx).Model(&rideModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"status": status, "updated_at": updatedAt}).Error
}
