package db

import "time"

type User struct {
	ID           string    `gorm:"column:id;type:uuid;primaryKey"`
	Email        *string   `gorm:"column:email"`
	Phone        *string   `gorm:"column:phone"`
	PasswordHash string    `gorm:"column:password_hash"`
	Role         string    `gorm:"column:role"`
	Name         string    `gorm:"column:name"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (User) TableName() string { return "users" }

type RiderProfile struct {
	UserID    string    `gorm:"column:user_id;type:uuid;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (RiderProfile) TableName() string { return "rider_profiles" }

type DriverProfile struct {
	UserID    string    `gorm:"column:user_id;type:uuid;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (DriverProfile) TableName() string { return "driver_profiles" }

type RefreshToken struct {
	ID        string     `gorm:"column:id;type:uuid;primaryKey"`
	UserID    string     `gorm:"column:user_id;type:uuid"`
	TokenHash string     `gorm:"column:token_hash"`
	ExpiresAt time.Time  `gorm:"column:expires_at"`
	RevokedAt *time.Time `gorm:"column:revoked_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }
