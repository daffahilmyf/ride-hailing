package app

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/adapters/cache"
	grpcadapter "github.com/daffahilmyf/ride-hailing/services/gateway/internal/adapters/grpc"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/handlers"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/middleware"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/infra"
)

func NewRouter(cfg infra.Config, logger *zap.Logger, deps Deps, redisClient *redis.Client, grpcClients *grpcadapter.Clients) *gin.Engine {
	r := gin.New()

	globalLimiter := cache.NewRedisLimiter(redisClient,
		cache.WithLimiterRequests(cfg.RateLimit.Requests),
		cache.WithLimiterWindow(time.Duration(cfg.RateLimit.WindowSeconds)*time.Second),
		cache.WithLimiterPrefix("gateway:rl"),
	)
	nearbyLimiter := cache.NewRedisLimiter(redisClient,
		cache.WithLimiterRequests(cfg.RateLimit.NearbyRequests),
		cache.WithLimiterWindow(time.Duration(cfg.RateLimit.NearbyWindowSeconds)*time.Second),
		cache.WithLimiterPrefix("gateway:nearby"),
	)
	notifyLimiter := cache.NewRedisLimiter(redisClient,
		cache.WithLimiterRequests(cfg.RateLimit.NotifyRequests),
		cache.WithLimiterWindow(time.Duration(cfg.RateLimit.NotifyWindowSeconds)*time.Second),
		cache.WithLimiterPrefix("gateway:notify"),
	)

	metrics := middleware.NewMetrics(cfg.ServiceName)
	r.Use(middleware.LoggerMiddleware(logger, cfg.ServiceName))
	r.Use(middleware.TraceMiddleware(cfg.ServiceName))
	r.Use(middleware.MetricsMiddleware(metrics))
	r.Use(middleware.RateLimitMiddleware(globalLimiter, cfg.RateLimit.Requests))
	r.Use(middleware.MaxBodyBytes(cfg.MaxBodyBytes))
	r.Use(middleware.RequestTimeout(time.Duration(cfg.HTTP.RequestTimeoutSeconds) * time.Second))

	r.StaticFile("/favicon.ico", "config/favicon.ico")
	var grpcConns []*grpc.ClientConn
	if grpcClients != nil {
		grpcConns = []*grpc.ClientConn{grpcClients.RideConn, grpcClients.MatchingConn, grpcClients.LocationConn}
	}

	readyCache := handlers.ReadinessCache{
		Cache: cache.NewRedisCache(redisClient),
		TTL:   cache.DefaultTTL(cfg.Cache),
		Key:   "gateway:readyz",
	}

	r.GET("/readyz", handlers.Ready(handlers.Readiness{
		Redis: redisClient,
		GRPC:  grpcConns,
		Cache: readyCache,
	}))

	v1 := r.Group("/v1")
	{
		v1.GET("/healthz", handlers.Health())
		v1.GET("/readyz", handlers.Ready(handlers.Readiness{
			Redis: redisClient,
			GRPC:  grpcConns,
			Cache: readyCache,
		}))
		v1.POST("/auth/register", handlers.ProxyUser(cfg.User.BaseURL, false, cfg.User.InternalToken))
		v1.POST("/auth/login", handlers.ProxyUser(cfg.User.BaseURL, false, cfg.User.InternalToken))
		v1.POST("/auth/refresh", handlers.ProxyUser(cfg.User.BaseURL, false, cfg.User.InternalToken))
		v1.POST("/auth/logout", handlers.ProxyUser(cfg.User.BaseURL, false, cfg.User.InternalToken))
		v1.POST("/auth/verify", handlers.ProxyUser(cfg.User.BaseURL, false, cfg.User.InternalToken))

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
		riderGroup.Use(middleware.AuditLogger(logger, "rides:write"))
		riderGroup.POST("/rides", handlers.CreateRide(deps.RideClient, cfg.GRPC.InternalToken))
		riderGroup.POST("/rides/:ride_id/cancel", handlers.CancelRide(deps.RideClient, cfg.GRPC.InternalToken))
		riderGroup.POST("/rides/:ride_id/offers", handlers.CreateOffer(deps.RideClient, cfg.GRPC.InternalToken))

		driverGroup := authGroup.Group("/")
		driverGroup.Use(middleware.RequireRole(middleware.RoleDriver))
		driverGroup.Use(middleware.RequireScope("drivers:write"))
		driverGroup.Use(middleware.AuditLogger(logger, "drivers:write"))
		driverGroup.POST("/drivers/:driver_id/status",
			middleware.RequireSubjectMatch("driver_id"),
			handlers.UpdateDriverStatus(deps.MatchingClient),
		)
		driverGroup.POST("/drivers/:driver_id/location",
			middleware.RequireSubjectMatch("driver_id"),
			handlers.UpdateDriverLocation(deps.LocationClient, cfg.GRPC.InternalToken),
		)
		driverGroup.POST("/drivers/nearby",
			middleware.RateLimitMiddleware(nearbyLimiter, cfg.RateLimit.NearbyRequests),
			handlers.ListNearbyDrivers(deps.LocationClient, cfg.GRPC.InternalToken),
		)
		driverGroup.POST("/offers/:offer_id/accept", handlers.AcceptOffer(deps.RideClient, cfg.GRPC.InternalToken))
		driverGroup.POST("/offers/:offer_id/decline", handlers.DeclineOffer(deps.RideClient, cfg.GRPC.InternalToken))
		driverGroup.POST("/offers/:offer_id/expire", handlers.ExpireOffer(deps.RideClient, cfg.GRPC.InternalToken))

		userGroup := authGroup.Group("/")
		userGroup.Use(middleware.RequireRole(middleware.RoleRider, middleware.RoleDriver))
		userGroup.GET("/users/me", handlers.ProxyUser(cfg.User.BaseURL, true, cfg.User.InternalToken))
		userGroup.POST("/auth/logout_all", handlers.ProxyUser(cfg.User.BaseURL, true, cfg.User.InternalToken))

		authGroup.GET("/notify/sse",
			middleware.RateLimitMiddleware(notifyLimiter, cfg.RateLimit.NotifyRequests),
			middleware.RequireScope("notify:read"),
			handlers.StreamNotifications(cfg.Notify.BaseURL),
		)
	}

	return r
}
