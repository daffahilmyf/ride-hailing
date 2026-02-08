package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userv1 "github.com/daffahilmyf/ride-hailing/proto/user/v1"
	grpcadapter "github.com/daffahilmyf/ride-hailing/services/gateway/internal/adapters/grpc"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/contextdata"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/handlers/requests"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/ports/outbound"
)

func RegisterAuth(authClient outbound.AuthService, internalToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}

		resp, err := authClient.Register(buildAuthCtx(c, internalToken), &userv1.RegisterRequest{
			Email:     strings.TrimSpace(req.Email),
			Phone:     strings.TrimSpace(req.Phone),
			Password:  req.Password,
			Role:      req.Role,
			Name:      req.Name,
			DeviceId:  deviceID(c, req.DeviceID),
			UserAgent: c.GetHeader("User-Agent"),
			Ip:        c.ClientIP(),
			TraceId:   contextdata.GetTraceID(c),
			RequestId: contextdata.GetRequestID(c),
		})
		if err != nil {
			respondAuthError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user":              mapUser(resp.GetUser()),
			"tokens":            mapTokens(resp.GetTokens()),
			"verification_code": resp.GetVerificationCode(),
		})
	}
}

func LoginAuth(authClient outbound.AuthService, internalToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}

		resp, err := authClient.Login(buildAuthCtx(c, internalToken), &userv1.LoginRequest{
			Email:     strings.TrimSpace(req.Email),
			Phone:     strings.TrimSpace(req.Phone),
			Password:  req.Password,
			DeviceId:  deviceID(c, req.DeviceID),
			UserAgent: c.GetHeader("User-Agent"),
			Ip:        c.ClientIP(),
			TraceId:   contextdata.GetTraceID(c),
			RequestId: contextdata.GetRequestID(c),
		})
		if err != nil {
			respondAuthError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user":   mapUser(resp.GetUser()),
			"tokens": mapTokens(resp.GetTokens()),
		})
	}
}

func RefreshAuth(authClient outbound.AuthService, internalToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.RefreshRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}

		resp, err := authClient.Refresh(buildAuthCtx(c, internalToken), &userv1.RefreshRequest{
			RefreshToken: req.RefreshToken,
			DeviceId:     deviceID(c, req.DeviceID),
			UserAgent:    c.GetHeader("User-Agent"),
			Ip:           c.ClientIP(),
			TraceId:      contextdata.GetTraceID(c),
			RequestId:    contextdata.GetRequestID(c),
		})
		if err != nil {
			respondAuthError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user":   mapUser(resp.GetUser()),
			"tokens": mapTokens(resp.GetTokens()),
		})
	}
}

func LogoutAuth(authClient outbound.AuthService, internalToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.LogoutRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}

		resp, err := authClient.Logout(buildAuthCtx(c, internalToken), &userv1.LogoutRequest{
			RefreshToken: req.RefreshToken,
			DeviceId:     deviceID(c, req.DeviceID),
			TraceId:      contextdata.GetTraceID(c),
			RequestId:    contextdata.GetRequestID(c),
		})
		if err != nil {
			respondAuthError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": resp.GetStatus()})
	}
}

func VerifyAuth(authClient outbound.AuthService, internalToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.VerifyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}

		resp, err := authClient.Verify(buildAuthCtx(c, internalToken), &userv1.VerifyRequest{
			Channel:   req.Channel,
			Target:    req.Target,
			Code:      req.Code,
			TraceId:   contextdata.GetTraceID(c),
			RequestId: contextdata.GetRequestID(c),
		})
		if err != nil {
			respondAuthError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": resp.GetStatus()})
	}
}

func LogoutAllAuth(authClient outbound.AuthService, internalToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := contextdata.GetUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		resp, err := authClient.LogoutAll(buildAuthCtx(c, internalToken), &userv1.LogoutAllRequest{
			UserId:    userID,
			TraceId:   contextdata.GetTraceID(c),
			RequestId: contextdata.GetRequestID(c),
		})
		if err != nil {
			respondAuthError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": resp.GetStatus()})
	}
}

func LogoutDeviceAuth(authClient outbound.AuthService, internalToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := contextdata.GetUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		var req requests.LogoutDeviceRequest
		if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.DeviceID) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}

		resp, err := authClient.LogoutDevice(buildAuthCtx(c, internalToken), &userv1.LogoutDeviceRequest{
			UserId:    userID,
			DeviceId:  strings.TrimSpace(req.DeviceID),
			TraceId:   contextdata.GetTraceID(c),
			RequestId: contextdata.GetRequestID(c),
		})
		if err != nil {
			respondAuthError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": resp.GetStatus()})
	}
}

func ListSessionsAuth(authClient outbound.AuthService, internalToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := contextdata.GetUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		resp, err := authClient.ListSessions(buildAuthCtx(c, internalToken), &userv1.ListSessionsRequest{
			UserId:    userID,
			TraceId:   contextdata.GetTraceID(c),
			RequestId: contextdata.GetRequestID(c),
		})
		if err != nil {
			respondAuthError(c, err)
			return
		}

		out := make([]gin.H, 0, len(resp.GetSessions()))
		for _, session := range resp.GetSessions() {
			out = append(out, gin.H{
				"device_id":  session.GetDeviceId(),
				"user_agent": session.GetUserAgent(),
				"ip":         session.GetIp(),
				"created_at": session.GetCreatedAt(),
				"expires_at": session.GetExpiresAt(),
			})
		}
		c.JSON(http.StatusOK, gin.H{"sessions": out})
	}
}

func MeAuth(authClient outbound.AuthService, internalToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := contextdata.GetUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		resp, err := authClient.GetMe(buildAuthCtx(c, internalToken), &userv1.GetMeRequest{
			UserId:    userID,
			TraceId:   contextdata.GetTraceID(c),
			RequestId: contextdata.GetRequestID(c),
		})
		if err != nil {
			respondAuthError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"user": mapUser(resp.GetUser())})
	}
}

func buildAuthCtx(c *gin.Context, internalToken string) context.Context {
	ctx := grpcadapter.WithRequestMetadata(
		c.Request.Context(),
		contextdata.GetTraceID(c),
		contextdata.GetRequestID(c),
	)
	ctx = grpcadapter.WithInternalToken(ctx, internalToken)
	ctx = grpcadapter.WithTraceContext(ctx)
	return ctx
}

func respondAuthError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}

	msg := st.Message()
	if msg == "" {
		msg = "internal"
	}

	switch st.Code() {
	case codes.InvalidArgument:
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
	case codes.AlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": msg})
	case codes.Unauthenticated:
		c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
	case codes.ResourceExhausted:
		c.JSON(http.StatusTooManyRequests, gin.H{"error": msg})
	case codes.NotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": msg})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
	}
}

func mapTokens(tokens *userv1.Token) gin.H {
	if tokens == nil {
		return gin.H{}
	}
	return gin.H{
		"access_token":  tokens.GetAccessToken(),
		"refresh_token": tokens.GetRefreshToken(),
		"token_type":    tokens.GetTokenType(),
		"expires_in":    tokens.GetExpiresIn(),
	}
}

func mapUser(user *userv1.User) gin.H {
	if user == nil {
		return gin.H{}
	}
	out := gin.H{
		"id":   user.GetId(),
		"role": user.GetRole(),
		"name": user.GetName(),
	}
	if user.GetEmail() != "" {
		out["email"] = user.GetEmail()
	}
	if user.GetPhone() != "" {
		out["phone"] = user.GetPhone()
	}
	if user.GetEmailVerifiedAt() != "" {
		out["email_verified_at"] = user.GetEmailVerifiedAt()
	}
	if user.GetPhoneVerifiedAt() != "" {
		out["phone_verified_at"] = user.GetPhoneVerifiedAt()
	}
	if user.GetRiderProfile() != nil {
		out["rider_profile"] = gin.H{
			"rating":             user.GetRiderProfile().GetRating(),
			"preferred_language": user.GetRiderProfile().GetPreferredLanguage(),
		}
	}
	if user.GetDriverProfile() != nil {
		out["driver_profile"] = gin.H{
			"vehicle_make":   user.GetDriverProfile().GetVehicleMake(),
			"vehicle_plate":  user.GetDriverProfile().GetVehiclePlate(),
			"license_number": user.GetDriverProfile().GetLicenseNumber(),
			"verified":       user.GetDriverProfile().GetVerified(),
			"rating":         user.GetDriverProfile().GetRating(),
		}
	}
	return out
}

func deviceID(c *gin.Context, fallback string) string {
	if v := c.GetHeader("X-Device-Id"); v != "" {
		return v
	}
	return strings.TrimSpace(fallback)
}
