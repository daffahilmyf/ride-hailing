package db

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type Repo struct {
	DB *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{DB: db}
}

func (r *Repo) CreateUserWithProfile(ctx context.Context, user User, role string) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		switch role {
		case "rider":
			profile := RiderProfile{UserID: user.ID, CreatedAt: time.Now().UTC()}
			return tx.Create(&profile).Error
		case "driver":
			profile := DriverProfile{UserID: user.ID, CreatedAt: time.Now().UTC()}
			return tx.Create(&profile).Error
		default:
			return errors.New("invalid role")
		}
	})
}

func (r *Repo) GetUserByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := r.DB.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return User{}, ErrNotFound
	}
	return user, err
}

func (r *Repo) GetUserByPhone(ctx context.Context, phone string) (User, error) {
	var user User
	err := r.DB.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return User{}, ErrNotFound
	}
	return user, err
}

func (r *Repo) GetUserByID(ctx context.Context, id string) (User, error) {
	var user User
	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return User{}, ErrNotFound
	}
	return user, err
}

func (r *Repo) CreateRefreshToken(ctx context.Context, token RefreshToken) error {
	return r.DB.WithContext(ctx).Create(&token).Error
}

func (r *Repo) GetRefreshToken(ctx context.Context, tokenHash string) (RefreshToken, error) {
	var token RefreshToken
	err := r.DB.WithContext(ctx).
		Where("token_hash = ? AND revoked_at IS NULL AND expires_at > NOW()", tokenHash).
		First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return RefreshToken{}, ErrNotFound
	}
	return token, err
}

func (r *Repo) GetRefreshTokenForDevice(ctx context.Context, tokenHash string, deviceID string) (RefreshToken, error) {
	var token RefreshToken
	err := r.DB.WithContext(ctx).
		Where("token_hash = ? AND device_id = ? AND revoked_at IS NULL AND expires_at > NOW()", tokenHash, deviceID).
		First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return RefreshToken{}, ErrNotFound
	}
	return token, err
}

func (r *Repo) GetRefreshTokenAny(ctx context.Context, tokenHash string) (RefreshToken, error) {
	var token RefreshToken
	err := r.DB.WithContext(ctx).
		Where("token_hash = ?", tokenHash).
		First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return RefreshToken{}, ErrNotFound
	}
	return token, err
}

func (r *Repo) RevokeRefreshToken(ctx context.Context, id string) error {
	now := time.Now().UTC()
	return r.DB.WithContext(ctx).
		Model(&RefreshToken{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", now).Error
}

func (r *Repo) RevokeAllRefreshTokens(ctx context.Context, userID string) error {
	now := time.Now().UTC()
	return r.DB.WithContext(ctx).
		Model(&RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}

func (r *Repo) HasActiveDeviceSession(ctx context.Context, userID string, deviceID string) (bool, error) {
	if userID == "" || deviceID == "" {
		return false, nil
	}
	var count int64
	err := r.DB.WithContext(ctx).
		Model(&RefreshToken{}).
		Where("user_id = ? AND device_id = ? AND revoked_at IS NULL AND expires_at > NOW()", userID, deviceID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repo) RevokeDeviceSessions(ctx context.Context, userID string, deviceID string) error {
	if userID == "" || deviceID == "" {
		return nil
	}
	now := time.Now().UTC()
	return r.DB.WithContext(ctx).
		Model(&RefreshToken{}).
		Where("user_id = ? AND device_id = ? AND revoked_at IS NULL", userID, deviceID).
		Update("revoked_at", now).Error
}

func (r *Repo) ListActiveDeviceSessions(ctx context.Context, userID string) ([]RefreshToken, error) {
	if userID == "" {
		return []RefreshToken{}, nil
	}
	var tokens []RefreshToken
	err := r.DB.WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > NOW()", userID).
		Order("created_at DESC").
		Find(&tokens).Error
	return tokens, err
}
