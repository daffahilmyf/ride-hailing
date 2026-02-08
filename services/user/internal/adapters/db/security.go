package db

import (
	"context"
	"time"

	"gorm.io/gorm"
)

func (r *Repo) IncrementFailedLogin(ctx context.Context, userID string, lockAfter int, lockDuration time.Duration) error {
	if userID == "" {
		return nil
	}
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user User
		if err := tx.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
			return err
		}
		count := user.FailedLoginCount + 1
		updates := map[string]any{"failed_login_count": count}
		if lockAfter > 0 && count >= lockAfter {
			until := time.Now().UTC().Add(lockDuration)
			updates["locked_until"] = until
		}
		return tx.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Updates(updates).Error
	})
}

func (r *Repo) ResetFailedLogin(ctx context.Context, userID string) error {
	if userID == "" {
		return nil
	}
	return r.DB.WithContext(ctx).
		Model(&User{}).
		Where("id = ?", userID).
		Updates(map[string]any{"failed_login_count": 0, "locked_until": nil}).Error
}
