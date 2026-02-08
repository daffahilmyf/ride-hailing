package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/daffahilmyf/ride-hailing/services/user/internal/adapters/db"
	"github.com/daffahilmyf/ride-hailing/services/user/internal/app/usecase"
)

type Handler struct {
	Service *usecase.Service
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

func RegisterRoutes(r *gin.Engine, svc *usecase.Service) {
	h := &Handler{Service: svc}
	v1 := r.Group("/v1")
	v1.POST("/auth/register", h.Register)
	v1.POST("/auth/login", h.Login)
	v1.POST("/auth/refresh", h.Refresh)
	v1.GET("/users/me", h.Me)
	v1.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
}

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		case usecase.ErrAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": "already_exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		}
		return
	}
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
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	user, tokens, err := h.Service.Login(c.Request.Context(), usecase.LoginInput{
		Email:    strings.TrimSpace(req.Email),
		Phone:    strings.TrimSpace(req.Phone),
		Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
		return
	}
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
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	user, tokens, err := h.Service.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_refresh"})
		return
	}
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

func (h *Handler) Me(c *gin.Context) {
	userID := c.GetHeader("X-User-Id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, err := h.Service.GetUser(c.Request.Context(), userID)
	if err != nil {
		if err == db.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": userResponse(user)})
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
