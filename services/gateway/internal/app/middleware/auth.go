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
	Enabled    bool
	JWTSecret  string
	JWTSecrets []string
	Issuer     string
	Audience   string
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
			secrets := effectiveSecrets(cfg)
			if len(secrets) == 0 {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secrets[0]), nil
		})
		if err != nil || !token.Valid {
			valid := false
			for _, secret := range effectiveSecrets(cfg) {
				t, e := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
					if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
						return nil, jwt.ErrSignatureInvalid
					}
					return []byte(secret), nil
				})
				if e == nil && t != nil && t.Valid {
					token = t
					valid = true
					break
				}
			}
			if !valid {
				LogGeneralError(c, logger, "auth.invalid_token", err)
				responses.RespondErrorCode(c, responses.CodeUnauthorized, map[string]string{"reason": "INVALID_TOKEN"})
				c.Abort()
				return
			}
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
		if scopes, ok := claims["scopes"]; ok {
			c.Set(scopesClaim, scopes)
		}
		contextdata.SetUserContext(c, userID, role)

		c.Next()
	}
}

func effectiveSecrets(cfg AuthConfig) []string {
	if len(cfg.JWTSecrets) > 0 {
		return cfg.JWTSecrets
	}
	if cfg.JWTSecret != "" {
		return []string{cfg.JWTSecret}
	}
	return nil
}
