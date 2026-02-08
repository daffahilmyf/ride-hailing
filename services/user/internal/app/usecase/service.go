package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/daffahilmyf/ride-hailing/services/user/internal/adapters/db"
	"github.com/daffahilmyf/ride-hailing/services/user/internal/infra"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidRole        = errors.New("invalid role")
	ErrAlreadyExists      = errors.New("already exists")
	ErrAccountLocked      = errors.New("account locked")
	ErrWeakPassword       = errors.New("weak password")
	ErrDeviceRequired     = errors.New("device required")
	ErrDeviceActive       = errors.New("device already active")
)

type Service struct {
	Repo       *db.Repo
	AuthConfig infra.AuthConfig
	Now        func() time.Time
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type RegisterInput struct {
	Email     string
	Phone     string
	Password  string
	Role      string
	Name      string
	DeviceID  string
	UserAgent string
	IP        string
}

type LoginInput struct {
	Email     string
	Phone     string
	Password  string
	DeviceID  string
	UserAgent string
	IP        string
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (db.User, Tokens, string, error) {
	role := strings.ToLower(strings.TrimSpace(in.Role))
	if role != "rider" && role != "driver" {
		return db.User{}, Tokens{}, "", ErrInvalidRole
	}
	if in.Email == "" && in.Phone == "" {
		return db.User{}, Tokens{}, "", ErrInvalidCredentials
	}
	if !validPassword(in.Password) {
		return db.User{}, Tokens{}, "", ErrWeakPassword
	}
	if in.DeviceID == "" {
		return db.User{}, Tokens{}, "", ErrDeviceRequired
	}

	if in.Email != "" {
		if _, err := s.Repo.GetUserByEmail(ctx, strings.ToLower(in.Email)); err == nil {
			return db.User{}, Tokens{}, "", ErrAlreadyExists
		}
	}
	if in.Phone != "" {
		if _, err := s.Repo.GetUserByPhone(ctx, in.Phone); err == nil {
			return db.User{}, Tokens{}, "", ErrAlreadyExists
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return db.User{}, Tokens{}, "", err
	}
	id := uuid.NewString()
	now := s.now()
	var email *string
	var phone *string
	if in.Email != "" {
		em := strings.ToLower(in.Email)
		email = &em
	}
	if in.Phone != "" {
		ph := in.Phone
		phone = &ph
	}
	user := db.User{
		ID:               id,
		Email:            email,
		Phone:            phone,
		PasswordHash:     string(hash),
		Role:             role,
		Name:             in.Name,
		FailedLoginCount: 0,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := s.Repo.CreateUserWithProfile(ctx, user, role); err != nil {
		return db.User{}, Tokens{}, "", err
	}
	code, _ := s.issueVerification(ctx, user)
	tokens, err := s.issueTokens(ctx, user, in.DeviceID, in.UserAgent, in.IP)
	if err != nil {
		return db.User{}, Tokens{}, "", err
	}
	return user, tokens, code, nil
}

func (s *Service) Login(ctx context.Context, in LoginInput) (db.User, Tokens, error) {
	if in.Email == "" && in.Phone == "" {
		return db.User{}, Tokens{}, ErrInvalidCredentials
	}
	if in.DeviceID == "" {
		return db.User{}, Tokens{}, ErrDeviceRequired
	}
	var user db.User
	var err error
	if in.Email != "" {
		user, err = s.Repo.GetUserByEmail(ctx, strings.ToLower(in.Email))
	} else {
		user, err = s.Repo.GetUserByPhone(ctx, in.Phone)
	}
	if err != nil {
		return db.User{}, Tokens{}, ErrInvalidCredentials
	}
	if user.LockedUntil != nil && user.LockedUntil.After(s.now()) {
		return db.User{}, Tokens{}, ErrAccountLocked
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)) != nil {
		_ = s.Repo.IncrementFailedLogin(ctx, user.ID, 5, 15*time.Minute)
		return db.User{}, Tokens{}, ErrInvalidCredentials
	}
	_ = s.Repo.ResetFailedLogin(ctx, user.ID)
	if user.Role == "driver" {
		_ = s.Repo.RevokeAllRefreshTokens(ctx, user.ID)
	} else {
		_ = s.Repo.RevokeDeviceSessions(ctx, user.ID, in.DeviceID)
	}
	tokens, err := s.issueTokens(ctx, user, in.DeviceID, in.UserAgent, in.IP)
	if err != nil {
		return db.User{}, Tokens{}, err
	}
	return user, tokens, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string, deviceID string) (db.User, Tokens, error) {
	if refreshToken == "" {
		return db.User{}, Tokens{}, ErrInvalidCredentials
	}
	if deviceID == "" {
		return db.User{}, Tokens{}, ErrDeviceRequired
	}
	hash := hashToken(refreshToken)
	rt, err := s.Repo.GetRefreshTokenForDevice(ctx, hash, deviceID)
	if err != nil {
		if any, anyErr := s.Repo.GetRefreshTokenAny(ctx, hash); anyErr == nil && any.RevokedAt != nil {
			_ = s.Repo.RevokeAllRefreshTokens(ctx, any.UserID)
		}
		if any, anyErr := s.Repo.GetRefreshTokenAny(ctx, hash); anyErr == nil && any.DeviceID != deviceID {
			_ = s.Repo.RevokeAllRefreshTokens(ctx, any.UserID)
		}
		return db.User{}, Tokens{}, ErrInvalidCredentials
	}
	user, err := s.Repo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return db.User{}, Tokens{}, ErrInvalidCredentials
	}
	_ = s.Repo.RevokeRefreshToken(ctx, rt.ID)
	tokens, err := s.issueTokens(ctx, user, deviceID, "", "")
	if err != nil {
		return db.User{}, Tokens{}, err
	}
	return user, tokens, nil
}

func (s *Service) GetUser(ctx context.Context, userID string) (db.User, error) {
	if userID == "" {
		return db.User{}, ErrInvalidCredentials
	}
	return s.Repo.GetUserByID(ctx, userID)
}

func (s *Service) Logout(ctx context.Context, refreshToken string, deviceID string) error {
	if refreshToken == "" {
		return ErrInvalidCredentials
	}
	if deviceID == "" {
		return ErrDeviceRequired
	}
	hash := hashToken(refreshToken)
	rt, err := s.Repo.GetRefreshTokenForDevice(ctx, hash, deviceID)
	if err != nil {
		return ErrInvalidCredentials
	}
	return s.Repo.RevokeRefreshToken(ctx, rt.ID)
}

func (s *Service) LogoutAll(ctx context.Context, userID string) error {
	if userID == "" {
		return ErrInvalidCredentials
	}
	return s.Repo.RevokeAllRefreshTokens(ctx, userID)
}

func (s *Service) Verify(ctx context.Context, channel string, target string, code string) error {
	if channel == "" || target == "" || code == "" {
		return ErrInvalidCredentials
	}
	hash := hashToken(code)
	v, err := s.Repo.ConsumeVerification(ctx, hash, channel)
	if err != nil {
		return ErrInvalidCredentials
	}
	user, err := s.Repo.GetUserByID(ctx, v.UserID)
	if err != nil {
		return ErrInvalidCredentials
	}
	if channel == "email" {
		if user.Email == nil || *user.Email != strings.ToLower(target) {
			return ErrInvalidCredentials
		}
		return s.Repo.MarkEmailVerified(ctx, v.UserID)
	}
	if channel == "phone" {
		if user.Phone == nil || *user.Phone != target {
			return ErrInvalidCredentials
		}
		return s.Repo.MarkPhoneVerified(ctx, v.UserID)
	}
	return ErrInvalidCredentials
}

func (s *Service) issueTokens(ctx context.Context, user db.User, deviceID string, userAgent string, ip string) (Tokens, error) {
	now := s.now()
	accessTTL := time.Duration(s.AuthConfig.AccessTTLSeconds) * time.Second
	if accessTTL <= 0 {
		accessTTL = 30 * time.Minute
	}
	refreshTTL := time.Duration(s.AuthConfig.RefreshTTLSeconds) * time.Second
	if refreshTTL <= 0 {
		refreshTTL = 30 * 24 * time.Hour
	}
	accessToken, err := s.signJWT(user, now.Add(accessTTL))
	if err != nil {
		return Tokens{}, err
	}
	refreshToken, refreshHash, err := newRefreshToken()
	if err != nil {
		return Tokens{}, err
	}
	rt := db.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		DeviceID:  deviceID,
		UserAgent: userAgent,
		IP:        ip,
		TokenHash: refreshHash,
		ExpiresAt: now.Add(refreshTTL),
		CreatedAt: now,
	}
	if err := s.Repo.CreateRefreshToken(ctx, rt); err != nil {
		return Tokens{}, err
	}
	return Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTTL.Seconds()),
	}, nil
}

func (s *Service) issueVerification(ctx context.Context, user db.User) (string, error) {
	code, hash, err := newVerificationCode()
	if err != nil {
		return "", err
	}
	channel := "email"
	if user.Phone != nil {
		channel = "phone"
	}
	v := db.VerificationCode{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		Channel:   channel,
		CodeHash:  hash,
		ExpiresAt: s.now().Add(10 * time.Minute),
		CreatedAt: s.now(),
	}
	if err := s.Repo.CreateVerification(ctx, v); err != nil {
		return "", err
	}
	return code, nil
}

func (s *Service) signJWT(user db.User, exp time.Time) (string, error) {
	secret := s.AuthConfig.JWTSecret
	if secret == "" {
		return "", errors.New("jwt secret required")
	}
	scopes := []string{"notify:read", "users:read"}
	if user.Role == "rider" {
		scopes = append(scopes, "rides:write")
	}
	if user.Role == "driver" {
		scopes = append(scopes, "drivers:write")
	}
	claims := jwt.MapClaims{
		"sub":    user.ID,
		"role":   user.Role,
		"scopes": scopes,
		"iss":    s.AuthConfig.Issuer,
		"aud":    s.AuthConfig.Audience,
		"exp":    exp.Unix(),
		"iat":    s.now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func (s *Service) now() time.Time {
	if s != nil && s.Now != nil {
		return s.Now()
	}
	return time.Now().UTC()
}

func validPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		switch {
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= '0' && r <= '9':
			hasDigit = true
		}
	}
	return hasUpper && hasLower && hasDigit
}

func newRefreshToken() (string, string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(buf)
	return token, hashToken(token), nil
}

func newVerificationCode() (string, string, error) {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	code := base64.RawURLEncoding.EncodeToString(buf)[:8]
	return code, hashToken(code), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
