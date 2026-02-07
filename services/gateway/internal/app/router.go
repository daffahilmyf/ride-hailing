package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/adapters/cache"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/handlers"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/middleware"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/infra"
	"time"
)

func NewRouter(cfg infra.Config, logger *zap.Logger, deps Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.LoggerMiddleware(logger, cfg.ServiceName))
	r.Use(middleware.MaxBodyBytes(cfg.MaxBodyBytes))
	limiter := cache.NewRedisLimiter(
		cache.NewRedisClient(cache.RedisConfig{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		}),
		cache.WithLimiterRequests(cfg.RateLimit.Requests),
		cache.WithLimiterWindow(time.Duration(cfg.RateLimit.WindowSeconds)*time.Second),
		cache.WithLimiterPrefix("rl"),
	)
	r.Use(middleware.RateLimitMiddleware(limiter))

	r.GET("/healthz", handlers.Health())

	v1 := r.Group("/v1")
	{
		v1.GET("/healthz", handlers.Health())

		authGroup := v1.Group("/")
		authGroup.Use(middleware.AuthMiddleware(logger, middleware.AuthConfig{
			Enabled:   cfg.Auth.Enabled,
			JWTSecret: cfg.Auth.JWTSecret,
			Issuer:    cfg.Auth.Issuer,
			Audience:  cfg.Auth.Audience,
		}))

		riderGroup := authGroup.Group("/")
		riderGroup.Use(middleware.RequireRole(middleware.RoleRider))
		riderGroup.Use(middleware.RequireScope("rides:write"))
		riderGroup.POST("/rides", handlers.CreateRide(deps.RideClient))
		riderGroup.POST("/rides/:ride_id/cancel", handlers.CancelRide(deps.RideClient))

		driverGroup := authGroup.Group("/")
		driverGroup.Use(middleware.RequireRole(middleware.RoleDriver))
		driverGroup.Use(middleware.RequireScope("drivers:write"))
		driverGroup.POST("/drivers/:driver_id/status", handlers.UpdateDriverStatus(deps.MatchingClient))
		driverGroup.POST("/drivers/:driver_id/location", handlers.UpdateDriverLocation(deps.LocationClient))
	}

	return r
}
