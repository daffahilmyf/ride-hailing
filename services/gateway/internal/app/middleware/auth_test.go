package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap/zaptest"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/contextdata"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)

	makeToken := func(secret string, claims jwt.MapClaims) string {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString([]byte(secret))
		if err != nil {
			t.Fatalf("sign token: %v", err)
		}
		return signed
	}

	tests := []struct {
		name       string
		cfg        AuthConfig
		authHeader string
		wantCode   int
		wantUser   string
	}{
		{
			name:     "disabled",
			cfg:      AuthConfig{Enabled: false},
			wantCode: http.StatusOK,
		},
		{
			name:     "missing_token",
			cfg:      AuthConfig{Enabled: true, JWTSecret: "secret"},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:       "invalid_token",
			cfg:        AuthConfig{Enabled: true, JWTSecret: "secret"},
			authHeader: "Bearer invalid",
			wantCode:   http.StatusUnauthorized,
		},
		{
			name: "valid_token",
			cfg: AuthConfig{
				Enabled:   true,
				JWTSecret: "secret",
				Issuer:    "issuer",
				Audience:  "aud",
			},
			authHeader: "Bearer " + makeToken("secret", jwt.MapClaims{
				"sub":   "user-1",
				"role":  "rider",
				"iss":   "issuer",
				"aud":   []string{"aud"},
				"exp":   time.Now().Add(time.Minute).Unix(),
				"scopes": []string{"ride:read"},
			}),
			wantCode: http.StatusOK,
			wantUser: "user-1",
		},
		{
			name: "rotation",
			cfg: AuthConfig{
				Enabled:    true,
				JWTSecrets: []string{"old", "new"},
			},
			authHeader: "Bearer " + makeToken("new", jwt.MapClaims{
				"sub": "user-2",
				"exp": time.Now().Add(time.Minute).Unix(),
			}),
			wantCode: http.StatusOK,
			wantUser: "user-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(AuthMiddleware(logger, tt.cfg))
			r.GET("/ping", func(c *gin.Context) {
				c.Header("X-User-ID", contextdata.GetUserID(c))
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/ping", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			r.ServeHTTP(w, req)
			if w.Code != tt.wantCode {
				t.Fatalf("expected %d, got %d", tt.wantCode, w.Code)
			}
			if tt.wantUser != "" {
				if got := w.Header().Get("X-User-ID"); got != tt.wantUser {
					t.Fatalf("expected user %q, got %q", tt.wantUser, got)
				}
			}
		})
	}
}
