package db

import "time"

type User struct {
	ID               string     `gorm:"column:id;type:uuid;primaryKey"`
	Email            *string    `gorm:"column:email"`
	Phone            *string    `gorm:"column:phone"`
	EmailVerifiedAt  *time.Time `gorm:"column:email_verified_at"`
	PhoneVerifiedAt  *time.Time `gorm:"column:phone_verified_at"`
	PasswordHash     string     `gorm:"column:password_hash"`
	Role             string     `gorm:"column:role"`
	Name             string     `gorm:"column:name"`
	FailedLoginCount int        `gorm:"column:failed_login_count"`
	LockedUntil      *time.Time `gorm:"column:locked_until"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

func (User) TableName() string { return "users" }

type RiderProfile struct {
	UserID            string    `gorm:"column:user_id;type:uuid;primaryKey"`
	Rating            float64   `gorm:"column:rating"`
	PreferredLanguage string    `gorm:"column:preferred_language"`
	CreatedAt         time.Time `gorm:"column:created_at"`
}

func (RiderProfile) TableName() string { return "rider_profiles" }

type DriverProfile struct {
	UserID        string    `gorm:"column:user_id;type:uuid;primaryKey"`
	VehicleMake   string    `gorm:"column:vehicle_make"`
	VehiclePlate  string    `gorm:"column:vehicle_plate"`
	LicenseNumber string    `gorm:"column:license_number"`
	Verified      bool      `gorm:"column:verified"`
	Rating        float64   `gorm:"column:rating"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (DriverProfile) TableName() string { return "driver_profiles" }

type RefreshToken struct {
	ID        string     `gorm:"column:id;type:uuid;primaryKey"`
	UserID    string     `gorm:"column:user_id;type:uuid"`
	DeviceID  string     `gorm:"column:device_id"`
	UserAgent string     `gorm:"column:user_agent"`
	IP        string     `gorm:"column:ip"`
	TokenHash string     `gorm:"column:token_hash"`
	ExpiresAt time.Time  `gorm:"column:expires_at"`
	RevokedAt *time.Time `gorm:"column:revoked_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

type VerificationCode struct {
	ID         string     `gorm:"column:id;type:uuid;primaryKey"`
	UserID     string     `gorm:"column:user_id;type:uuid"`
	Channel    string     `gorm:"column:channel"`
	CodeHash   string     `gorm:"column:code_hash"`
	ExpiresAt  time.Time  `gorm:"column:expires_at"`
	ConsumedAt *time.Time `gorm:"column:consumed_at"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
}

func (VerificationCode) TableName() string { return "verification_codes" }
