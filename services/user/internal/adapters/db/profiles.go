package db

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

func (r *Repo) GetRiderProfile(ctx context.Context, userID string) (RiderProfile, error) {
	var prof RiderProfile
	err := r.DB.WithContext(ctx).Where("user_id = ?", userID).First(&prof).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return RiderProfile{}, ErrNotFound
	}
	return prof, err
}

func (r *Repo) GetDriverProfile(ctx context.Context, userID string) (DriverProfile, error) {
	var prof DriverProfile
	err := r.DB.WithContext(ctx).Where("user_id = ?", userID).First(&prof).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DriverProfile{}, ErrNotFound
	}
	return prof, err
}

func (r *Repo) CreateVerification(ctx context.Context, v VerificationCode) error {
	return r.DB.WithContext(ctx).Create(&v).Error
}

func (r *Repo) ConsumeVerification(ctx context.Context, codeHash string, channel string) (VerificationCode, error) {
	var v VerificationCode
	err := r.DB.WithContext(ctx).
		Where("code_hash = ? AND channel = ? AND consumed_at IS NULL AND expires_at > NOW()", codeHash, channel).
		First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return VerificationCode{}, ErrNotFound
	}
	now := time.Now().UTC()
	if err := r.DB.WithContext(ctx).Model(&VerificationCode{}).
		Where("id = ? AND consumed_at IS NULL", v.ID).
		Update("consumed_at", now).Error; err != nil {
		return VerificationCode{}, err
	}
	v.ConsumedAt = &now
	return v, nil
}

func (r *Repo) MarkEmailVerified(ctx context.Context, userID string) error {
	now := time.Now().UTC()
	return r.DB.WithContext(ctx).
		Model(&User{}).
		Where("id = ?", userID).
		Update("email_verified_at", now).Error
}

func (r *Repo) MarkPhoneVerified(ctx context.Context, userID string) error {
	now := time.Now().UTC()
	return r.DB.WithContext(ctx).
		Model(&User{}).
		Where("id = ?", userID).
		Update("phone_verified_at", now).Error
}
