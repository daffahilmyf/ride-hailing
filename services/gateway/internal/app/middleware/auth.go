package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/contextdata"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/responses"
)

type AuthConfig struct {
	Enabled   bool
	JWTSecret string
	Issuer    string
	Audience  string
}

func AuthMiddleware(logger *zap.Logger, cfg AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enabled {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			responses.RespondErrorCode(c, responses.CodeUnauthorized, map[string]string{"reason": "MISSING_TOKEN"})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			LogGeneralError(c, logger, "auth.invalid_token", err)
			responses.RespondErrorCode(c, responses.CodeUnauthorized, map[string]string{"reason": "INVALID_TOKEN"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			responses.RespondErrorCode(c, responses.CodeUnauthorized, map[string]string{"reason": "INVALID_CLAIMS"})
			c.Abort()
			return
		}

		if cfg.Issuer != "" {
			issuer, _ := claims.GetIssuer()
			if issuer != cfg.Issuer {
				responses.RespondErrorCode(c, responses.CodeUnauthorized, map[string]string{"reason": "INVALID_ISSUER"})
				c.Abort()
				return
			}
		}

		if cfg.Audience != "" {
			if aud, _ := claims.GetAudience(); len(aud) == 0 || aud[0] != cfg.Audience {
				responses.RespondErrorCode(c, responses.CodeUnauthorized, map[string]string{"reason": "INVALID_AUDIENCE"})
				c.Abort()
				return
			}
		}

		userID := ""
		if sub, _ := claims.GetSubject(); sub != "" {
			userID = sub
		}
		role, _ := claims["role"].(string)
		contextdata.SetUserContext(c, userID, role)

		c.Next()
	}
}
