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
}

type loginRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

func RegisterRoutes(r *gin.Engine, svc *usecase.Service, logger *zap.Logger, authMetrics *metrics.AuthMetrics, limiter *RateLimiter, internalAuthEnabled bool, internalAuthToken string) {
	h := &Handler{Service: svc, Logger: logger, Metrics: authMetrics}
	v1 := r.Group("/v1")

	if limiter != nil {
		v1.POST("/auth/register", RateLimitMiddleware(limiter), h.Register)
		v1.POST("/auth/login", RateLimitMiddleware(limiter), h.Login)
		v1.POST("/auth/refresh", RateLimitMiddleware(limiter), h.Refresh)
		v1.POST("/auth/logout", RateLimitMiddleware(limiter), h.Logout)
	} else {
		v1.POST("/auth/register", h.Register)
		v1.POST("/auth/login", h.Login)
		v1.POST("/auth/refresh", h.Refresh)
		v1.POST("/auth/logout", h.Logout)
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
	user, tokens, err := h.Service.Register(c.Request.Context(), usecase.RegisterInput{
		Email:    strings.TrimSpace(req.Email),
		Phone:    strings.TrimSpace(req.Phone),
		Password: req.Password,
		Role:     req.Role,
		Name:     req.Name,
	})
	if err != nil {
		switch err {
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
	h.audit("register", user.ID, "ok")
	h.respondWithMetrics(c, "register", "ok", start)
	c.JSON(http.StatusOK, gin.H{
		"user": userResponse(user),
		"tokens": tokenResponse{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			TokenType:    "bearer",
			ExpiresIn:    tokens.ExpiresIn,
		},
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
		Email:    strings.TrimSpace(req.Email),
		Phone:    strings.TrimSpace(req.Phone),
		Password: req.Password,
	})
	if err != nil {
		h.audit("login", "", "invalid_credentials")
		h.respondWithMetrics(c, "login", "unauthorized", start)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
		return
	}
	h.audit("login", user.ID, "ok")
	h.respondWithMetrics(c, "login", "ok", start)
	c.JSON(http.StatusOK, gin.H{
		"user": userResponse(user),
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
	user, tokens, err := h.Service.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.respondWithMetrics(c, "refresh", "unauthorized", start)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_refresh"})
		return
	}
	h.audit("refresh", user.ID, "ok")
	h.respondWithMetrics(c, "refresh", "ok", start)
	c.JSON(http.StatusOK, gin.H{
		"user": userResponse(user),
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
	if err := h.Service.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		h.respondWithMetrics(c, "logout", "unauthorized", start)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_refresh"})
		return
	}
	h.audit("logout", "", "ok")
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
	h.audit("logout_all", userID, "ok")
	h.respondWithMetrics(c, "logout_all", "ok", start)
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
	h.respondWithMetrics(c, "me", "ok", start)
	c.JSON(http.StatusOK, gin.H{"user": userResponse(user)})
}

func (h *Handler) respondWithMetrics(_ *gin.Context, endpoint string, status string, start time.Time) {
	if h.Metrics != nil {
		h.Metrics.Record(endpoint, status, time.Since(start))
	}
}

func (h *Handler) audit(action string, userID string, status string) {
	if h.Logger == nil {
		return
	}
	h.Logger.Info("audit",
		zap.String("action", action),
		zap.String("user_id", userID),
		zap.String("status", status),
	)
}

func userResponse(user db.User) gin.H {
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
	return resp
}
