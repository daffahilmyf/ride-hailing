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
	Email    string
	Phone    string
	Password string
	Role     string
	Name     string
}

type LoginInput struct {
	Email    string
	Phone    string
	Password string
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (db.User, Tokens, error) {
	role := strings.ToLower(strings.TrimSpace(in.Role))
	if role != "rider" && role != "driver" {
		return db.User{}, Tokens{}, ErrInvalidRole
	}
	if in.Email == "" && in.Phone == "" {
		return db.User{}, Tokens{}, ErrInvalidCredentials
	}
	if in.Password == "" {
		return db.User{}, Tokens{}, ErrInvalidCredentials
	}

	if in.Email != "" {
		if _, err := s.Repo.GetUserByEmail(ctx, strings.ToLower(in.Email)); err == nil {
			return db.User{}, Tokens{}, ErrAlreadyExists
		}
	}
	if in.Phone != "" {
		if _, err := s.Repo.GetUserByPhone(ctx, in.Phone); err == nil {
			return db.User{}, Tokens{}, ErrAlreadyExists
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return db.User{}, Tokens{}, err
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
		ID:           id,
		Email:        email,
		Phone:        phone,
		PasswordHash: string(hash),
		Role:         role,
		Name:         in.Name,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.Repo.CreateUserWithProfile(ctx, user, role); err != nil {
		return db.User{}, Tokens{}, err
	}
	tokens, err := s.issueTokens(ctx, user)
	if err != nil {
		return db.User{}, Tokens{}, err
	}
	return user, tokens, nil
}

func (s *Service) Login(ctx context.Context, in LoginInput) (db.User, Tokens, error) {
	if in.Email == "" && in.Phone == "" {
		return db.User{}, Tokens{}, ErrInvalidCredentials
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
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)) != nil {
		return db.User{}, Tokens{}, ErrInvalidCredentials
	}
	tokens, err := s.issueTokens(ctx, user)
	if err != nil {
		return db.User{}, Tokens{}, err
	}
	return user, tokens, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (db.User, Tokens, error) {
	if refreshToken == "" {
		return db.User{}, Tokens{}, ErrInvalidCredentials
	}
	hash := hashToken(refreshToken)
	rt, err := s.Repo.GetRefreshToken(ctx, hash)
	if err != nil {
		return db.User{}, Tokens{}, ErrInvalidCredentials
	}
	user, err := s.Repo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return db.User{}, Tokens{}, ErrInvalidCredentials
	}
	_ = s.Repo.RevokeRefreshToken(ctx, rt.ID)
	tokens, err := s.issueTokens(ctx, user)
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

func (s *Service) issueTokens(ctx context.Context, user db.User) (Tokens, error) {
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

func newRefreshToken() (string, string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(buf)
	return token, hashToken(token), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
