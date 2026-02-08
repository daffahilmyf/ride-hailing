package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/daffahilmyf/ride-hailing/services/user/internal/adapters/db"
	"github.com/daffahilmyf/ride-hailing/services/user/internal/app/metrics"
	"github.com/daffahilmyf/ride-hailing/services/user/internal/app/usecase"
)

type Handler struct {
	Service *usecase.Service
	Logger  *zap.Logger
	Metrics *metrics.AuthMetrics
}

type registerRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Name     string `json:"name"`
	DeviceID string `json:"device_id"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	DeviceID string `json:"device_id"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
	DeviceID     string `json:"device_id"`
}

type verifyRequest struct {
	Channel string `json:"channel"`
	Target  string `json:"target"`
	Code    string `json:"code"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

func RegisterRoutes(r *gin.Engine, svc *usecase.Service, logger *zap.Logger, authMetrics *metrics.AuthMetrics, limiter *RateLimiter, internalAuthEnabled bool, internalAuthToken string) {
	h := &Handler{Service: svc, Logger: logger, Metrics: authMetrics}
	r.Use(RequestIDMiddleware())
	v1 := r.Group("/v1")

	if limiter != nil {
		v1.POST("/auth/register", RateLimitMiddleware(limiter), h.Register)
		v1.POST("/auth/login", RateLimitMiddleware(limiter), h.Login)
		v1.POST("/auth/refresh", RateLimitMiddleware(limiter), h.Refresh)
		v1.POST("/auth/logout", RateLimitMiddleware(limiter), h.Logout)
		v1.POST("/auth/verify", RateLimitMiddleware(limiter), h.Verify)
	} else {
		v1.POST("/auth/register", h.Register)
		v1.POST("/auth/login", h.Login)
		v1.POST("/auth/refresh", h.Refresh)
		v1.POST("/auth/logout", h.Logout)
		v1.POST("/auth/verify", h.Verify)
	}

	protected := v1.Group("/")
	protected.Use(InternalAuthMiddleware(internalAuthEnabled, internalAuthToken))
	protected.GET("/users/me", h.Me)
	protected.POST("/auth/logout_all", h.LogoutAll)

	v1.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
}

func (h *Handler) Register(c *gin.Context) {
	start := time.Now()
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithMetrics(c, "register", "bad_request", start)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	user, tokens, code, err := h.Service.Register(c.Request.Context(), usecase.RegisterInput{
		Email:     strings.TrimSpace(req.Email),
		Phone:     strings.TrimSpace(req.Phone),
		Password:  req.Password,
		Role:      req.Role,
		Name:      req.Name,
		DeviceID:  deviceID(c, req.DeviceID),
		UserAgent: c.GetHeader("User-Agent"),
		IP:        c.ClientIP(),
	})
	if err != nil {
		switch err {
		case usecase.ErrWeakPassword:
			h.respondWithMetrics(c, "register", "weak_password", start)
			c.JSON(http.StatusBadRequest, gin.H{"error": "weak_password"})
		case usecase.ErrInvalidCredentials, usecase.ErrInvalidRole:
			h.respondWithMetrics(c, "register", "bad_request", start)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		case usecase.ErrAlreadyExists:
			h.respondWithMetrics(c, "register", "conflict", start)
			c.JSON(http.StatusConflict, gin.H{"error": "already_exists"})
		default:
			h.respondWithMetrics(c, "register", "error", start)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		}
		return
	}
	h.audit(c, "register", user.ID, "ok")
	h.respondWithMetrics(c, "register", "ok", start)
	c.JSON(http.StatusOK, gin.H{
		"user": userResponse(user, nil, nil),
		"tokens": tokenResponse{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			TokenType:    "bearer",
			ExpiresIn:    tokens.ExpiresIn,
		},
		"verification_code": code,
	})
}

func (h *Handler) Login(c *gin.Context) {
	start := time.Now()
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithMetrics(c, "login", "bad_request", start)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	user, tokens, err := h.Service.Login(c.Request.Context(), usecase.LoginInput{
		Email:     strings.TrimSpace(req.Email),
		Phone:     strings.TrimSpace(req.Phone),
		Password:  req.Password,
		DeviceID:  deviceID(c, req.DeviceID),
		UserAgent: c.GetHeader("User-Agent"),
		IP:        c.ClientIP(),
	})
	if err != nil {
		if err == usecase.ErrAccountLocked {
			h.audit(c, "login", user.ID, "locked")
			h.respondWithMetrics(c, "login", "locked", start)
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "account_locked"})
			return
		}
		h.audit(c, "login", "", "invalid_credentials")
		h.respondWithMetrics(c, "login", "unauthorized", start)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
		return
	}
	h.audit(c, "login", user.ID, "ok")
	h.respondWithMetrics(c, "login", "ok", start)
	c.JSON(http.StatusOK, gin.H{
		"user": userResponse(user, nil, nil),
		"tokens": tokenResponse{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			TokenType:    "bearer",
			ExpiresIn:    tokens.ExpiresIn,
		},
	})
}

func (h *Handler) Refresh(c *gin.Context) {
	start := time.Now()
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithMetrics(c, "refresh", "bad_request", start)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	user, tokens, err := h.Service.Refresh(c.Request.Context(), req.RefreshToken, deviceID(c, req.DeviceID))
	if err != nil {
		h.respondWithMetrics(c, "refresh", "unauthorized", start)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_refresh"})
		return
	}
	h.audit(c, "refresh", user.ID, "ok")
	h.respondWithMetrics(c, "refresh", "ok", start)
	c.JSON(http.StatusOK, gin.H{
		"user": userResponse(user, nil, nil),
		"tokens": tokenResponse{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			TokenType:    "bearer",
			ExpiresIn:    tokens.ExpiresIn,
		},
	})
}

func (h *Handler) Logout(c *gin.Context) {
	start := time.Now()
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithMetrics(c, "logout", "bad_request", start)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	if err := h.Service.Logout(c.Request.Context(), req.RefreshToken, deviceID(c, req.DeviceID)); err != nil {
		h.respondWithMetrics(c, "logout", "unauthorized", start)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_refresh"})
		return
	}
	h.audit(c, "logout", "", "ok")
	h.respondWithMetrics(c, "logout", "ok", start)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) LogoutAll(c *gin.Context) {
	start := time.Now()
	userID := c.GetHeader("X-User-Id")
	if userID == "" {
		h.respondWithMetrics(c, "logout_all", "unauthorized", start)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if err := h.Service.LogoutAll(c.Request.Context(), userID); err != nil {
		h.respondWithMetrics(c, "logout_all", "error", start)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}
	h.audit(c, "logout_all", userID, "ok")
	h.respondWithMetrics(c, "logout_all", "ok", start)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Verify(c *gin.Context) {
	start := time.Now()
	var req verifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithMetrics(c, "verify", "bad_request", start)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	if err := h.Service.Verify(c.Request.Context(), req.Channel, req.Target, req.Code); err != nil {
		h.respondWithMetrics(c, "verify", "unauthorized", start)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_code"})
		return
	}
	h.audit(c, "verify", "", "ok")
	h.respondWithMetrics(c, "verify", "ok", start)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Me(c *gin.Context) {
	start := time.Now()
	userID := c.GetHeader("X-User-Id")
	if userID == "" {
		h.respondWithMetrics(c, "me", "unauthorized", start)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, err := h.Service.GetUser(c.Request.Context(), userID)
	if err != nil {
		if err == db.ErrNotFound {
			h.respondWithMetrics(c, "me", "not_found", start)
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		h.respondWithMetrics(c, "me", "error", start)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}
	var rider *db.RiderProfile
	var driver *db.DriverProfile
	if user.Role == "rider" {
		if prof, err := h.Service.Repo.GetRiderProfile(c.Request.Context(), user.ID); err == nil {
			rider = &prof
		}
	} else if user.Role == "driver" {
		if prof, err := h.Service.Repo.GetDriverProfile(c.Request.Context(), user.ID); err == nil {
			driver = &prof
		}
	}
	h.respondWithMetrics(c, "me", "ok", start)
	c.JSON(http.StatusOK, gin.H{"user": userResponse(user, rider, driver)})
}

func (h *Handler) respondWithMetrics(c *gin.Context, endpoint string, status string, start time.Time) {
	if h.Metrics != nil {
		h.Metrics.Record(endpoint, status, time.Since(start))
	}
	if h.Logger != nil {
		h.Logger.Debug("request", zap.String("endpoint", endpoint), zap.String("status", status), zap.String("trace_id", GetTraceID(c)), zap.String("request_id", GetRequestID(c)))
	}
}

func (h *Handler) audit(c *gin.Context, action string, userID string, status string) {
	if h.Logger == nil {
		return
	}
	h.Logger.Info("audit",
		zap.String("action", action),
		zap.String("user_id", userID),
		zap.String("status", status),
		zap.String("trace_id", GetTraceID(c)),
		zap.String("request_id", GetRequestID(c)),
	)
}

func userResponse(user db.User, rider *db.RiderProfile, driver *db.DriverProfile) gin.H {
	resp := gin.H{
		"id":   user.ID,
		"role": user.Role,
		"name": user.Name,
	}
	if user.Email != nil {
		resp["email"] = *user.Email
	}
	if user.Phone != nil {
		resp["phone"] = *user.Phone
	}
	if user.EmailVerifiedAt != nil {
		resp["email_verified_at"] = user.EmailVerifiedAt.UTC().Format(time.RFC3339)
	}
	if user.PhoneVerifiedAt != nil {
		resp["phone_verified_at"] = user.PhoneVerifiedAt.UTC().Format(time.RFC3339)
	}
	if rider != nil {
		resp["rider_profile"] = gin.H{
			"rating":             rider.Rating,
			"preferred_language": rider.PreferredLanguage,
		}
	}
	if driver != nil {
		resp["driver_profile"] = gin.H{
			"vehicle_make":   driver.VehicleMake,
			"vehicle_plate":  driver.VehiclePlate,
			"license_number": driver.LicenseNumber,
			"verified":       driver.Verified,
			"rating":         driver.Rating,
		}
	}
	return resp
}

func deviceID(c *gin.Context, fallback string) string {
	if v := c.GetHeader("X-Device-Id"); v != "" {
		return v
	}
	return strings.TrimSpace(fallback)
}
